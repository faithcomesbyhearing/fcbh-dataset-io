/**
 * Artie Main Application Logic
 * Handles folder selection, validation, and form population
 */

// Global state
let currentFolderData = null;
let currentFolderInfo = null;
let validationResult = null;

// Upload control
let uploadController = null; // For canceling uploads
let isUploading = false;

/**
 * Initialize folder upload functionality
 */
function initializeFolderUpload() {
    setupFolderDropzone();
    setupUploadButton();
}

/**
 * Setup folder dropzone (drag-and-drop + click-to-select)
 */
function setupFolderDropzone() {
    const dropzone = document.getElementById('folderDropzone');
    if (!dropzone) return;
    
    // Drag and drop events
    dropzone.addEventListener('dragover', function(e) {
        e.preventDefault();
        dropzone.classList.add('dragover');
    });
    
    dropzone.addEventListener('dragleave', function(e) {
        e.preventDefault();
        dropzone.classList.remove('dragover');
    });
    
    dropzone.addEventListener('drop', function(e) {
        e.preventDefault();
        dropzone.classList.remove('dragover');
        
        // Show immediate feedback
        showFolderStatus('📁 Processing dropped folder...', 'processing');
        
        const items = e.dataTransfer.items;
        if (items && items.length > 0) {
            // Check if it's a directory
            const item = items[0];
            if (item.kind === 'file') {
                // Try webkit API first (Chrome, Edge, Safari)
                if (item.webkitGetAsEntry) {
                    const entry = item.webkitGetAsEntry();
                    if (entry && entry.isDirectory) {
                        handleFolderSelection(entry);
                        return;
                    } else if (entry && entry.isFile) {
                        showFolderStatus('Please drop a folder, not a file', 'error');
                        return;
                    }
                }
                
                // Try File System Access API (if available)
                if (item.getAsFileSystemHandle) {
                    item.getAsFileSystemHandle().then(handle => {
                        if (handle.kind === 'directory') {
                            handleDirectoryPicker(handle);
                            return;
                        } else {
                            showFolderStatus('Please drop a folder, not a file', 'error');
                            return;
                        }
                    }).catch(err => {
                        console.error('File System Access API error:', err);
                        showFolderStatus('Folder access not supported in this browser', 'error');
                    });
                    return;
                }
                
                // Fallback message
                showFolderStatus('Folder drag-and-drop not supported. Please use "Click to select"', 'error');
            } else {
                showFolderStatus('Please drop a folder', 'error');
            }
        } else {
            showFolderStatus('No folder detected in drop', 'error');
        }
    });
    
    // Click to select folder
    dropzone.addEventListener('click', function() {
        if (window.showDirectoryPicker) {
            // Modern File System Access API
            showFolderStatus('📁 Please select a folder...', 'processing');
            window.showDirectoryPicker().then(handleDirectoryPicker)
                .catch(err => {
                    if (err.name !== 'AbortError') {
                        console.error('Directory picker error:', err);
                        showFolderStatus('Failed to select folder', 'error');
                    } else {
                        // User cancelled - reset to default state
                        showFolderStatus('No folder selected', 'default');
                    }
                });
        } else {
            // Fallback for older browsers
            showFolderStatus('📁 Please select a folder...', 'processing');
            const input = document.createElement('input');
            input.type = 'file';
            input.webkitdirectory = true;
            input.onchange = function(e) {
                if (e.target.files.length > 0) {
                    handleFileList(e.target.files);
                } else {
                    showFolderStatus('No folder selected', 'default');
                }
            };
            input.click();
        }
    });
}

/**
 * Handle directory picker selection (modern browsers)
 */
async function handleDirectoryPicker(directoryHandle) {
    try {
        const folderData = await readDirectoryContentsFSA(directoryHandle);
        const folderName = directoryHandle.name;
        processFolderData(folderData, folderName);
    } catch (error) {
        console.error('Error reading directory:', error);
        showFolderStatus('Failed to read folder contents', 'error');
    }
}

/**
 * Handle folder selection (called from drag-and-drop)
 */
async function handleFolderSelection(entry) {
    try {
        const folderData = await readDirectoryContentsWebkit(entry);
        const folderName = entry.name;
        processFolderData(folderData, folderName);
    } catch (error) {
        console.error('Error reading directory:', error);
        showFolderStatus('Failed to read folder contents', 'error');
    }
}

/**
 * Handle file list from webkitdirectory input (older browsers)
 */
function handleFileList(files) {
    const folderData = organizeFilesByType(files);
    const folderName = extractFolderNameFromPath(files[0]?.webkitRelativePath || '');
    processFolderData(folderData, folderName);
}

/**
 * Read directory contents using File System Access API (for showDirectoryPicker)
 */
async function readDirectoryContentsFSA(directoryHandle) {
    const files = [];
    
    for await (const [name, handle] of directoryHandle.entries()) {
        if (handle.kind === 'file') {
            const file = await handle.getFile();
            files.push(file);
        } else if (handle.kind === 'directory') {
            // Recursively read subdirectory
            const subFiles = await readDirectoryContentsFSA(handle);
            files.push(...subFiles);
        }
    }
    
    return organizeFilesByType(files);
}

/**
 * Read directory contents using webkit API (for drag-and-drop)
 */
async function readDirectoryContentsWebkit(directoryEntry, currentPath = '') {
    const allFiles = [];
    
    const readDirectory = async (entry, path = '', depth = 0) => {
        // Prevent infinite recursion (safety limit)
        if (depth > 10) {
            console.warn(`⚠️  Maximum recursion depth reached at: ${path}/${entry.name}`);
            return;
        }
        
        try {
            if (entry.isFile) {
                const file = await new Promise((resolve, reject) => {
                    entry.file(resolve, reject);
                });
                // Create our own relative path since webkitRelativePath is unreliable
                file.customRelativePath = path ? `${path}/${entry.name}` : entry.name;
                
                // No debug logging for file discovery
                allFiles.push(file);
            } else if (entry.isDirectory) {
                const newPath = path ? `${path}/${entry.name}` : entry.name;
                
                const reader = entry.createReader();
                
                // Read all entries from this directory with robust error handling
                const readAllEntries = async () => {
                    try {
                        const entries = await new Promise((resolve, reject) => {
                            reader.readEntries(resolve, reject);
                        });
                        
                        if (entries.length === 0) {
                            return; // No more entries
                        }
                        
                        // Process each entry with error handling
                        for (const subEntry of entries) {
                            try {
                                await readDirectory(subEntry, newPath, depth + 1); // Recursive call with updated path and depth
                            } catch (entryError) {
                                console.error(`❌ Error processing entry ${subEntry.name}:`, entryError);
                                // Continue processing other entries even if one fails
                            }
                        }
                        
                        // Continue reading more entries (readEntries may not return all at once)
                        await readAllEntries();
                    } catch (readError) {
                        console.error(`❌ Error reading directory ${entry.name}:`, readError);
                        // Continue processing even if this directory fails
                    }
                };
                
                await readAllEntries();
            }
        } catch (error) {
            console.error(`❌ Error processing ${entry.name}:`, error);
            // Continue processing other entries even if this one fails
        }
    };
    
    await readDirectory(directoryEntry, currentPath);
    
    return organizeFilesByType(allFiles);
}

/**
 * Organize files by type (audio, text)
 */
function organizeFilesByType(files) {
    const audioFiles = [];
    const textFiles = [];
    let audioSubfolder = '';
    let textSubfolder = '';
    
    // Organizing files by type
    
    for (const file of files) {
        // Use our custom relative path that we built during directory traversal
        const path = file.customRelativePath || file.webkitRelativePath || file.name;
        const pathParts = path.split('/');
        
        // Look for audio files (mp3, wav) in any subfolder
        if (file.name.match(/\.(mp3|wav)$/i)) {
            audioFiles.push(file);
            // Try to identify the audio subfolder (more robust detection)
            if (pathParts.length >= 2) {
                const subfolder = pathParts[pathParts.length - 2];
                // Look for common audio folder patterns
                if (!audioSubfolder || 
                    subfolder.includes('VOX') || 
                    subfolder.includes('Audio') || 
                    subfolder.includes('audio') ||
                    subfolder.includes('MP3') ||
                    subfolder.includes('mp3')) {
                    audioSubfolder = subfolder;
                }
            }
        }
        
        // Look for text files (usx) in any subfolder
        else if (file.name.toLowerCase().endsWith('.usx')) {
            textFiles.push(file);
            
            // Try to identify the text subfolder (more robust detection)
            if (pathParts.length >= 2) {
                const subfolder = pathParts[pathParts.length - 2];
                // Look for common text folder patterns
                if (!textSubfolder || 
                    subfolder.includes('USX') || 
                    subfolder.includes('usx') ||
                    subfolder.includes('Text') || 
                    subfolder.includes('text') ||
                    subfolder.includes('TXT') ||
                    subfolder.includes('txt')) {
                    textSubfolder = subfolder;
                }
            }
        }
    }
    
    // File organization complete
    
    
    return {
        audioFiles,
        textFiles,
        audioSubfolder: audioSubfolder || 'audio',
        textSubfolder: textSubfolder || 'text',
        totalFiles: files.length
    };
}

/**
 * Extract folder name from file path
 */
function extractFolderNameFromPath(relativePath) {
    const parts = relativePath.split('/');
    return parts[0] || 'unknown';
}

/**
 * Process folder data (validate and populate form)
 */
async function processFolderData(folderData, folderName) {
    // Add folder name to data
    folderData.folderName = folderName;
    
    // Show initial feedback
    showFolderStatus('📁 Analyzing folder contents...', 'processing');
    updateFolderProgress(10, 100, 'Reading folder structure');
    
    // Parse folder name
    updateFolderProgress(30, 100, 'Parsing folder name');
    const folderInfo = parseFolderName(folderName);
    if (!folderInfo.valid) {
        showFolderStatus(folderInfo.error, 'error');
        return;
    }
    
    // Validate folder structure
    updateFolderProgress(60, 100, 'Validating file structure');
    
    // Add a small delay to show the progress
    await new Promise(resolve => setTimeout(resolve, 200));
    
    const validation = validateFolderStructure(folderData);
    
    updateFolderProgress(90, 100, 'Finalizing validation');
    await new Promise(resolve => setTimeout(resolve, 100));
    
    if (validation.valid) {
        // Store data for upload
        currentFolderData = folderData;
        currentFolderInfo = folderInfo;
        validationResult = validation;
        
        // Populate form fields
        populateFormFromFolder(folderInfo, folderData);
        
        // Show success status
        showFolderStatus(`✅ Valid folder: ${validation.audioBooks.length} books, ${validation.totalAudioFiles} audio files, ${validation.totalTextFiles} text files`, 'success');
        
        // Enable upload button if credentials are loaded
        updateUploadButtonState();
        
    } else {
        // Show validation errors in modal
        showFolderStatus(`❌ Validation failed - click to see details`, 'error');
        showErrorModal(validation.errors);
        
        // Clear any previous data
        currentFolderData = null;
        currentFolderInfo = null;
        validationResult = null;
        updateUploadButtonState();
    }
}

/**
 * Populate form fields from folder data
 */
function populateFormFromFolder(folderInfo, folderData) {
    // Get bucket name
    const bucketName = getBucketName();
    
    // Populate basic fields
    document.getElementById('datasetName').value = folderInfo.datasetName;
    document.getElementById('languageIso').value = folderInfo.iso;
    
    // Populate S3 paths (using testing/ prefix)
    const textPath = `s3://${bucketName}/testing/uploaded/${folderData.folderName}/${folderData.textSubfolder}/*.usx`;
    const audioPath = `s3://${bucketName}/testing/uploaded/${folderData.folderName}/${folderData.audioSubfolder}/*.mp3`;
    
    document.getElementById('textData').value = textPath;
    document.getElementById('audioData').value = audioPath;
    
    // Clear any previous error styling
    clearFieldError('datasetName');
    clearFieldError('languageIso');
    clearFieldError('textData');
    clearFieldError('audioData');
}

/**
 * Setup upload button functionality
 */
function setupUploadButton() {
    const uploadButton = document.getElementById('uploadButton');
    if (!uploadButton) return;
    
    uploadButton.addEventListener('click', async function() {
        console.log('🚀 Upload button clicked!');
        
        // Handle cancel upload
        if (isUploading) {
            console.log('🛑 Canceling upload...');
            if (uploadController) {
                uploadController.abort();
            }
            showStatus('Upload cancelled', 'warning');
            return;
        }
        
        console.log('  - currentFolderData:', !!currentFolderData);
        console.log('  - currentFolderInfo:', !!currentFolderInfo);
        console.log('  - validationResult:', !!validationResult);
        console.log('  - AWS credentials:', !!AWS_CONFIG.accessKeyId && !!AWS_CONFIG.secretAccessKey);
        
        if (!currentFolderData || !currentFolderInfo || !validationResult) {
            console.log('❌ Missing folder data');
            showStatus('❌ Please select and validate a folder first', 'error');
            return;
        }
        
        if (!AWS_CONFIG.accessKeyId || !AWS_CONFIG.secretAccessKey) {
            console.log('❌ Missing AWS credentials');
            showStatus('❌ Please load AWS credentials first', 'error');
            return;
        }
        
        // Check username
        const username = document.getElementById('username').value.trim();
        console.log('🔍 Username check:', `"${username}"`, 'Length:', username.length);
        if (!username) {
            console.log('❌ Missing username - showing error');
            showStatus('❌ Please enter a username (required field)', 'error');
            // Highlight the username field
            const usernameField = document.getElementById('username');
            usernameField.classList.add('required-error');
            usernameField.classList.remove('required-valid');
            usernameField.focus();
            return;
        }
        
        console.log('✅ All checks passed, starting upload...');
        
        // Warn about same folder upload
        const folderName = currentFolderData.folderName;
        showStatus(`⚠️ Uploading folder: ${folderName} (will overwrite existing files)`, 'warning');
        
        // Set upload state
        isUploading = true;
        uploadController = new AbortController();
        
        // Update button for upload state
        uploadButton.disabled = false;
        uploadButton.textContent = 'Cancel Upload';
        uploadButton.style.backgroundColor = '#dc3545'; // Red for cancel
        
        try {
            const bucketName = getBucketName();
            
            // Show progress
            const progressDiv = document.getElementById('uploadProgress');
            if (progressDiv) {
                progressDiv.style.display = 'block';
            }
            
            // Perform upload
            const result = await performUpload(currentFolderData, currentFolderInfo, bucketName, 
                (current, total, message) => {
                    updateUploadProgress(current, total, message);
                },
                uploadController.signal
            );
            
            // Show success message
            showStatus(`✅ Upload complete! Files uploaded to s3://${bucketName}/testing/uploaded/${currentFolderData.folderName}/ and YAML saved to s3://${bucketName}/testing/input/${currentFolderInfo.datasetName}.yaml`, 'success');
            
            // Also save locally
            saveFile();
            
        } catch (error) {
            console.error('Upload error:', error);
            showStatus(`Upload failed: ${error.message}`, 'error');
        } finally {
            // Reset upload state
            isUploading = false;
            uploadController = null;
            
            // Hide progress bar
            const progressBar = document.getElementById('uploadProgressBar');
            if (progressBar) {
                progressBar.style.display = 'none';
            }
            
            // Reset button
            uploadButton.disabled = false;
            uploadButton.textContent = 'Upload';
            uploadButton.style.backgroundColor = ''; // Reset to default color
            
            // Hide progress
            const progressDiv = document.getElementById('uploadProgress');
            if (progressDiv) {
                progressDiv.style.display = 'none';
            }
        }
    });
}

/**
 * Update upload button state based on current conditions
 */
// Make function globally accessible
window.updateUploadButtonState = function() {
    const uploadButton = document.getElementById('uploadButton');
    if (!uploadButton) {
        console.log('❌ Upload button not found');
        return;
    }
    
    const hasCredentials = AWS_CONFIG.accessKeyId && AWS_CONFIG.secretAccessKey;
    const hasValidFolder = currentFolderData && currentFolderInfo && validationResult;
    const hasUsername = document.getElementById('username').value.trim().length > 0;
    
    console.log('🔍 Upload button state check:');
    console.log('  - Credentials loaded:', hasCredentials);
    console.log('  - Valid folder:', hasValidFolder);
    console.log('  - Username entered:', hasUsername);
    console.log('  - Username value:', document.getElementById('username').value);
    
    uploadButton.disabled = !hasCredentials || !hasValidFolder || !hasUsername;
    
    if (!hasCredentials) {
        uploadButton.title = 'Load AWS credentials first';
        console.log('  - Button disabled: Missing credentials');
    } else if (!hasValidFolder) {
        uploadButton.title = 'Select and validate a folder first';
        console.log('  - Button disabled: No valid folder');
    } else if (!hasUsername) {
        uploadButton.title = 'Enter a username (required field)';
        console.log('  - Button disabled: No username');
    } else {
        uploadButton.title = 'Upload folder to S3';
        console.log('  - Button ENABLED: All requirements met');
    }
};

/**
 * Show folder status message
 */
function showFolderStatus(message, type) {
    const statusDiv = document.getElementById('folderStatus');
    const dropzone = document.getElementById('folderDropzone');
    
    if (statusDiv) {
        statusDiv.textContent = message;
        statusDiv.className = `folder-status ${type}`;
    }
    
    if (dropzone) {
        dropzone.className = `folder-dropzone ${type}`;
    }
    
    // Hide progress bar when not processing
    if (type !== 'processing') {
        hideFolderProgress();
    }
}

/**
 * Update folder validation progress
 */
function updateFolderProgress(current, total, message) {
    const progressDiv = document.getElementById('folderProgress');
    if (!progressDiv) return;
    
    const percentage = Math.round((current / total) * 100);
    progressDiv.style.display = 'block';
    progressDiv.innerHTML = `
        <div class="progress-bar">
            <div class="progress-fill" style="width: ${percentage}%"></div>
        </div>
        <div class="progress-text">${message} (${percentage}%)</div>
    `;
}

/**
 * Hide folder progress indicator
 */
function hideFolderProgress() {
    const progressDiv = document.getElementById('folderProgress');
    if (progressDiv) {
        progressDiv.style.display = 'none';
    }
}

/**
 * Update upload progress display
 */
function updateUploadProgress(current, total, message) {
    const progressBar = document.getElementById('uploadProgressBar');
    const progressFill = document.getElementById('uploadProgressFill');
    const progressText = document.getElementById('uploadProgressText');
    const progressMessage = document.getElementById('uploadProgressMessage');
    
    if (progressBar && progressFill && progressText && progressMessage) {
        const percentage = Math.round((current / total) * 100);
        
        // Show progress bar if hidden
        if (progressBar.style.display === 'none') {
            progressBar.style.display = 'block';
        }
        
        // Check if upload is complete
        if (percentage === 100) {
            // Replace progress bar with celebratory completion message
            progressFill.style.width = `100%`;
            progressFill.style.backgroundColor = '#00ff00'; // Bright green
            progressText.textContent = `🎉 100% 🎉`;
            progressMessage.innerHTML = `🎊 <strong style="color: #00ff00; font-size: 16px;">Upload Completed!</strong> 🎊`;
        } else {
            // Normal progress update
            progressFill.style.width = `${percentage}%`;
            progressFill.style.backgroundColor = '#4CAF50'; // Normal green
            progressText.textContent = `${percentage}%`;
            
            // Show only the last uploaded file name (no file count)
            if (message) {
                progressMessage.textContent = `Last uploaded: ${message}`;
            } else {
                progressMessage.textContent = `Uploading...`;
            }
        }
    }
}

/**
 * Clear field error styling
 */
function clearFieldError(fieldId) {
    const element = document.getElementById(fieldId);
    if (element) {
        element.classList.remove('required-error');
    }
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    // Wait a bit for other scripts to load
    setTimeout(initializeFolderUpload, 100);
});
