/**
 * Artie Upload Functionality
 * Handles S3 upload with progress tracking and YAML generation
 */

/**
 * Check existing files in S3 bucket for smart re-upload
 */
async function checkExistingFiles(folderData, folderName, bucketName) {
    const s3 = new AWS.S3();
    const existingFiles = new Map(); // filename -> size
    
    try {
        // Check audio files
        const audioPrefix = `testing/uploaded/${folderName}/${folderData.audioSubfolder}/`;
        const audioList = await s3.listObjectsV2({
            Bucket: bucketName,
            Prefix: audioPrefix
        }).promise();
        
        if (audioList.Contents) {
            audioList.Contents.forEach(obj => {
                const filename = obj.Key.split('/').pop(); // Get just the filename
                existingFiles.set(filename, obj.Size);
            });
        }
        
        // Check text files
        const textPrefix = `testing/uploaded/${folderName}/${folderData.textSubfolder}/`;
        const textList = await s3.listObjectsV2({
            Bucket: bucketName,
            Prefix: textPrefix
        }).promise();
        
        if (textList.Contents) {
            textList.Contents.forEach(obj => {
                const filename = obj.Key.split('/').pop(); // Get just the filename
                existingFiles.set(filename, obj.Size);
            });
        }
        
        return existingFiles;
    } catch (error) {
        console.warn('Could not check existing files, will upload all:', error.message);
        return new Map(); // Return empty map if we can't check
    }
}

/**
 * Upload files to S3 with progress tracking and smart re-upload
 */
async function uploadFilesToS3(folderData, folderName, bucketName, onProgress, abortSignal = null) {
    const s3 = new AWS.S3();
    const uploadedFiles = [];
    const totalFiles = folderData.audioFiles.length + folderData.textFiles.length;
    let uploadedCount = 0;
    let skippedCount = 0;
    
    // Check existing files first
    onProgress(0, totalFiles, 'Checking existing files...');
    const existingFiles = await checkExistingFiles(folderData, folderName, bucketName);
    
    // Upload audio files
    for (const file of folderData.audioFiles) {
        // Check for abort signal
        if (abortSignal && abortSignal.aborted) {
            throw new Error('Upload cancelled');
        }
        
        const key = `testing/uploaded/${folderName}/${folderData.audioSubfolder}/${file.name}`;
        const existingSize = existingFiles.get(file.name);
        
        // Check if file already exists with same size
        if (existingSize !== undefined && existingSize === file.size) {
            // File exists with same size, skip upload
            skippedCount++;
            uploadedCount++;
            uploadedFiles.push(key); // Still track it as "processed"
            
            if (onProgress) {
                onProgress(uploadedCount, totalFiles, `Skipped: ${file.name} (already exists)`);
            }
            continue;
        }
        
        try {
            const fileContent = await readFileAsArrayBuffer(file);
            await s3.upload({
                Bucket: bucketName,
                Key: key,
                Body: fileContent,
                ContentType: getContentType(file.name)
            }).promise();
            
            uploadedFiles.push(key);
            uploadedCount++;
            
            if (onProgress) {
                onProgress(uploadedCount, totalFiles, `Uploaded: ${file.name}`);
            }
        } catch (error) {
            if (abortSignal && abortSignal.aborted) {
                throw new Error('Upload cancelled');
            }
            throw new Error(`Failed to upload audio file ${file.name}: ${error.message}`);
        }
    }
    
    // Upload text files
    for (const file of folderData.textFiles) {
        // Check for abort signal
        if (abortSignal && abortSignal.aborted) {
            throw new Error('Upload cancelled');
        }
        
        const key = `testing/uploaded/${folderName}/${folderData.textSubfolder}/${file.name}`;
        const existingSize = existingFiles.get(file.name);
        
        // Check if file already exists with same size
        if (existingSize !== undefined && existingSize === file.size) {
            // File exists with same size, skip upload
            skippedCount++;
            uploadedCount++;
            uploadedFiles.push(key); // Still track it as "processed"
            
            if (onProgress) {
                onProgress(uploadedCount, totalFiles, `Skipped: ${file.name} (already exists)`);
            }
            continue;
        }
        
        try {
            const fileContent = await readFileAsArrayBuffer(file);
            await s3.upload({
                Bucket: bucketName,
                Key: key,
                Body: fileContent,
                ContentType: 'text/xml' // USX files are XML
            }).promise();
            
            uploadedFiles.push(key);
            uploadedCount++;
            
            if (onProgress) {
                onProgress(uploadedCount, totalFiles, `Uploaded: ${file.name}`);
            }
        } catch (error) {
            if (abortSignal && abortSignal.aborted) {
                throw new Error('Upload cancelled');
            }
            throw new Error(`Failed to upload text file ${file.name}: ${error.message}`);
        }
    }
    
    return {
        uploadedFiles: uploadedFiles,
        uploadedCount: uploadedCount - skippedCount,
        skippedCount: skippedCount,
        totalProcessed: uploadedCount
    };
}

/**
 * Read file as ArrayBuffer (for binary files)
 */
function readFileAsArrayBuffer(file) {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = () => resolve(reader.result);
        reader.onerror = () => reject(new Error('Failed to read file'));
        reader.readAsArrayBuffer(file);
    });
}

/**
 * Get content type based on file extension
 */
function getContentType(filename) {
    const ext = filename.toLowerCase().split('.').pop();
    switch (ext) {
        case 'mp3': return 'audio/mpeg';
        case 'wav': return 'audio/wav';
        case 'usx': return 'text/xml';
        default: return 'application/octet-stream';
    }
}

/**
 * Generate YAML content for the request
 */
function generateUploadYAML(folderData, folderInfo, bucketName) {
    const datasetName = folderInfo.datasetName;
    const iso = folderInfo.iso;
    const audioPath = `s3://${bucketName}/uploaded/${folderData.folderName}/${folderData.audioSubfolder}/*.{mp3,wav}`;
    const textPath = `s3://${bucketName}/uploaded/${folderData.folderName}/${folderData.textSubfolder}/*.usx`;
    
    // Get current form values or use defaults
    const username = document.getElementById('username').value || 'uploaded_user';
    const notifyOk = document.getElementById('notifyOk').value || 'jbarndt@fcbhmail.org, ezornes@fcbhmail.org, gfiddes@fcbhmail.org, edomschot@fcbhmail.org';
    const notifyErr = document.getElementById('notifyErr').value || 'jbarndt@fcbhmail.org, gary@shortsands.com, ezornes@fcbhmail.org';
    const gordonFilter = document.getElementById('gordonFilter').value || '4';
    
    // Get radio button values
    const timestampsValue = document.querySelector('input[name="timestamps"]:checked')?.value || 'mms_align';
    const trainingValue = document.querySelector('input[name="training"]:checked')?.value || 'mms_adapter';
    const sttValue = document.querySelector('input[name="speech_to_text"]:checked')?.value || 'adapter_asr';
    const compareChecked = document.getElementById('compare').checked;
    
    // Parse email lists
    const notifyOkEmails = notifyOk.split(',').map(email => email.trim()).filter(email => email);
    const notifyErrEmails = notifyErr.split(',').map(email => email.trim()).filter(email => email);
    
    // Build YAML structure
    const yamlData = {
        is_new: true,
        dataset_name: datasetName,
        username: username,
        language_iso: iso,
        notify_ok: notifyOkEmails,
        notify_err: notifyErrEmails,
        text_data: {
            aws_s3: textPath
        },
        audio_data: {
            aws_s3: audioPath
        }
    };
    
    // Add timestamps
    yamlData.timestamps = {};
    if (timestampsValue === 'mms_align') yamlData.timestamps.mms_align = true;
    else if (timestampsValue === 'mms_fa_verse') yamlData.timestamps.mms_fa_verse = true;
    else yamlData.timestamps.no_timestamps = true;
    
    // Add training
    yamlData.training = {};
    if (trainingValue === 'mms_adapter') {
        yamlData.training.mms_adapter = {
            batch_mb: 4,
            num_epochs: 16,
            learning_rate: 0.001,
            warmup_pct: 12,
            grad_norm_max: 0.4
        };
    } else {
        yamlData.training.no_training = true;
    }
    
    // Add speech to text
    yamlData.speech_to_text = {};
    if (sttValue === 'mms_asr') yamlData.speech_to_text.mms_asr = true;
    else if (sttValue === 'adapter_asr') yamlData.speech_to_text.adapter_asr = true;
    else yamlData.speech_to_text.no_speech_to_text = true;
    
    // Add compare settings
    if (compareChecked) {
        yamlData.compare = {
            html_report: true,
            gordon_filter: parseInt(gordonFilter) || 4,
            compare_settings: {
                lower_case: true,
                remove_prompt_chars: true,
                remove_punctuation: true,
                double_quotes: { remove: true },
                apostrophe: { remove: true },
                hyphen: { remove: true },
                diacritical_marks: { normalize_nfc: true }
            }
        };
    }
    
    return yamlData;
}

/**
 * Upload YAML file to S3
 */
async function uploadYAMLToS3(yamlData, datasetName, bucketName) {
    const s3 = new AWS.S3();
    
    // Convert to YAML string
    let yamlString;
    if (window.jsyaml) {
        yamlString = window.jsyaml.dump(yamlData, { indent: 4, lineWidth: -1 });
    } else {
        // Fallback to simple YAML generation
        yamlString = generateSimpleYAML(yamlData);
    }
    
    const key = `testing/input/${datasetName}.yaml`;
    
    await s3.upload({
        Bucket: bucketName,
        Key: key,
        Body: yamlString,
        ContentType: 'text/yaml'
    }).promise();
    
    return key;
}

/**
 * Simple YAML generator (fallback when js-yaml is not available)
 */
function generateSimpleYAML(obj) {
    let yaml = '';
    
    function processValue(key, value, indent = 0) {
        const spaces = '    '.repeat(indent);
        
        if (value === null || value === undefined) return '';
        if (typeof value === 'boolean') return `${spaces}${key}: ${value}`;
        if (typeof value === 'number') return `${spaces}${key}: ${value}`;
        if (typeof value === 'string') return `${spaces}${key}: ${value}`;
        
        if (Array.isArray(value)) {
            let result = `${spaces}${key}:\n`;
            value.forEach(item => {
                result += `${spaces}    - ${item}\n`;
            });
            return result;
        }
        
        if (typeof value === 'object') {
            let result = `${spaces}${key}:\n`;
            for (const [k, v] of Object.entries(value)) {
                result += processValue(k, v, indent + 1) + '\n';
            }
            return result;
        }
        
        return `${spaces}${key}: ${value}`;
    }
    
    for (const [key, value] of Object.entries(obj)) {
        yaml += processValue(key, value) + '\n';
    }
    
    return yaml;
}

/**
 * Main upload function
 */
async function performUpload(folderData, folderInfo, bucketName, onProgress, abortSignal = null) {
    try {
        // Step 1: Upload files
        onProgress(0, 100, 'Starting file upload...');
        const uploadResult = await uploadFilesToS3(folderData, folderData.folderName, bucketName, 
            (current, total, message) => {
                const percentage = Math.round((current / total) * 80); // Files upload = 80% of progress
                onProgress(percentage, 100, message);
            },
            abortSignal
        );
        
        // Step 2: Generate and upload YAML
        onProgress(85, 100, 'Generating YAML...');
        const yamlData = generateUploadYAML(folderData, folderInfo, bucketName);
        
        onProgress(90, 100, 'Uploading YAML...');
        const yamlKey = await uploadYAMLToS3(yamlData, folderInfo.datasetName, bucketName);
        
        onProgress(100, 100, 'Upload complete!');
        
        return {
            success: true,
            uploadedFiles: uploadResult.uploadedFiles,
            yamlKey: yamlKey,
            uploadedCount: uploadResult.uploadedCount,
            skippedCount: uploadResult.skippedCount,
            totalProcessed: uploadResult.totalProcessed,
            message: `Successfully processed ${uploadResult.totalProcessed} files (${uploadResult.uploadedCount} uploaded, ${uploadResult.skippedCount} skipped) and YAML configuration`
        };
        
    } catch (error) {
        throw new Error(`Upload failed: ${error.message}`);
    }
}
