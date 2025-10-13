#!/bin/bash

# Configure CORS for S3 buckets to allow local development
# Usage: ./configure-s3-cors.sh

AUDIO_BUCKET="pretest-audio"
YAML_BUCKET="dataset-io"

echo "Configuring CORS for S3 buckets: $AUDIO_BUCKET and $YAML_BUCKET"

# Check if AWS CLI is installed
if ! command -v aws &> /dev/null; then
    echo "❌ AWS CLI is not installed. Please install it first."
    echo "   Visit: https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html"
    exit 1
fi

# Check if credentials are configured
if ! aws sts get-caller-identity &> /dev/null; then
    echo "❌ AWS credentials not configured. Please run 'aws configure' first."
    exit 1
fi

# Apply CORS configuration to audio bucket
echo "📋 Applying CORS configuration to $AUDIO_BUCKET..."
aws s3api put-bucket-cors --bucket "$AUDIO_BUCKET" --cors-configuration file://s3-cors-config.json

if [ $? -eq 0 ]; then
    echo "✅ CORS configuration applied successfully to $AUDIO_BUCKET!"
else
    echo "❌ Failed to apply CORS configuration to $AUDIO_BUCKET."
    echo "   Make sure you have permissions to modify bucket CORS settings."
    exit 1
fi

# Apply CORS configuration to YAML bucket
echo "📋 Applying CORS configuration to $YAML_BUCKET..."
aws s3api put-bucket-cors --bucket "$YAML_BUCKET" --cors-configuration file://s3-cors-config.json

if [ $? -eq 0 ]; then
    echo "✅ CORS configuration applied successfully to $YAML_BUCKET!"
else
    echo "❌ Failed to apply CORS configuration to $YAML_BUCKET."
    echo "   Make sure you have permissions to modify bucket CORS settings."
    exit 1
fi

echo ""
echo "🔍 You can verify the configurations with:"
echo "   aws s3api get-bucket-cors --bucket $AUDIO_BUCKET"
echo "   aws s3api get-bucket-cors --bucket $YAML_BUCKET"
echo ""
echo "🚀 Your web app should now be able to upload files to both S3 buckets."
