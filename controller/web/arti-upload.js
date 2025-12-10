/**
 * Artie Upload Functionality
 * Handles S3 upload with progress tracking and YAML generation
 */

/**
 * Check existing files in S3 bucket for smart re-upload
 */
async function checkExistingFiles(folderData, folderName, audioBucketName) {
    const s3 = new AWS.S3();
    const existingFiles = new Map(); // filename -> size
    
    try {
        // Check audio files
        const audioPrefix = `${folderName}/${folderData.audioSubfolder}/`;
        const audioList = await s3.listObjectsV2({
            Bucket: audioBucketName,
            Prefix: audioPrefix
        }).promise();
        
        if (audioList.Contents) {
            audioList.Contents.forEach(obj => {
                const filename = obj.Key.split('/').pop(); // Get just the filename
                existingFiles.set(filename, obj.Size);
            });
        }
        
        // Check text files
        const textPrefix = `${folderName}/${folderData.textSubfolder}/`;
        const textList = await s3.listObjectsV2({
            Bucket: audioBucketName,
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
async function uploadFilesToS3(folderData, folderName, audioBucketName, onProgress, abortSignal = null) {
    const s3 = new AWS.S3();
    const uploadedFiles = [];
    const totalFiles = folderData.audioFiles.length + folderData.textFiles.length;
    let uploadedCount = 0;
    let skippedCount = 0;
    
    // Check existing files first
    onProgress(0, totalFiles, 'Checking existing files...');
    const existingFiles = await checkExistingFiles(folderData, folderName, audioBucketName);
    
    // Upload audio files
    for (const file of folderData.audioFiles) {
        // Check for abort signal
        if (abortSignal && abortSignal.aborted) {
            throw new Error('Upload cancelled');
        }
        
        const key = `${folderName}/${folderData.audioSubfolder}/${file.name}`;
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
                Bucket: audioBucketName,
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
        
        const key = `${folderName}/${folderData.textSubfolder}/${file.name}`;
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
                Bucket: audioBucketName,
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
 * Uses the main buildYAMLObject() function and overrides the paths for folder uploads
 */
function generateUploadYAML(folderData, folderInfo, audioBucketName) {
    // Use the main buildYAMLObject function to get all form values as an object
    const yamlData = window.buildYAMLObject();
    if (!yamlData) {
        return null;
    }
    
    // Override the paths for folder uploads
    const audioPath = `s3://${audioBucketName}/${folderData.folderName}/${folderData.audioSubfolder}/*.mp3`;
    const textPath = `s3://${audioBucketName}/${folderData.folderName}/${folderData.textSubfolder}/*.usx`;
    
    yamlData.text_data = { aws_s3: textPath };
    yamlData.audio_data = { aws_s3: audioPath };
    
    return yamlData;
}

/**
 * Upload YAML file to S3
 */
async function uploadYAMLToS3(yamlData, datasetName, yamlBucketName) {
    const s3 = new AWS.S3();
    
    // Convert to YAML string
    let yamlString;
    if (window.jsyaml) {
        yamlString = window.jsyaml.dump(yamlData, { indent: 4, lineWidth: -1 });
    } else {
        // Fallback to simple YAML generation
        yamlString = generateSimpleYAML(yamlData);
    }
    
    // Post-process to ensure csv: yes is unquoted
    yamlString = yamlString.replace(/^(\s+)csv:\s*['"]yes['"]/m, '$1csv: yes');
    
    const key = `input/${datasetName}.yaml`;
    
    await s3.upload({
        Bucket: yamlBucketName,
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
async function performUpload(folderData, folderInfo, audioBucketName, yamlBucketName, onProgress, abortSignal = null) {
    try {
        // Step 1: Upload files
        onProgress(0, 100, 'Starting file upload...');
        const uploadResult = await uploadFilesToS3(folderData, folderData.folderName, audioBucketName, 
            (current, total, message) => {
                const percentage = Math.round((current / total) * 80); // Files upload = 80% of progress
                onProgress(percentage, 100, message);
            },
            abortSignal
        );
        
        // Step 2: Generate and upload YAML
        onProgress(85, 100, 'Generating YAML...');
        const yamlData = generateUploadYAML(folderData, folderInfo, audioBucketName);
        
        onProgress(90, 100, 'Uploading YAML...');
        const currentDatasetName = document.getElementById('datasetName').value || folderInfo.datasetName;
        const yamlKey = await uploadYAMLToS3(yamlData, currentDatasetName, yamlBucketName);
        
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
