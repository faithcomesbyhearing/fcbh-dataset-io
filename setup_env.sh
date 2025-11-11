#!/bin/bash

# FCBH Dataset I/O Environment Setup for Mac
# Run this script with: source setup_env.sh

# Set the GOPROJ environment variable to the current directory
export GOPROJ=$(pwd)

# Set database and file paths
export FCBH_DATASET_DB="$HOME/tmp/artie/db"
export FCBH_DATASET_FILES="$HOME/tmp/artie/files"
export FCBH_DATASET_TMP="$HOME/tmp/artie/tmp"

# Create directories if they don't exist
mkdir -p "$FCBH_DATASET_DB"
mkdir -p "$FCBH_DATASET_FILES"
mkdir -p "$FCBH_DATASET_TMP"

# Verify system dependencies
echo "Checking system dependencies..."
if ! command -v ffmpeg &> /dev/null; then
    echo "❌ FFmpeg not found. Install with: brew install ffmpeg (Mac) or sudo apt install ffmpeg (Ubuntu)"
    exit 1
fi

if ! command -v sox &> /dev/null; then
    echo "❌ Sox not found. Install with: brew install sox (Mac) or sudo apt install sox (Ubuntu)"
    exit 1
fi

echo "✅ System dependencies verified"
echo ""

# Logging configuration
# Option 1 (Recommended for production): Per-job logs in a directory
export FCBH_DATASET_LOG_DIR="$HOME/tmp/artie/logs"
# Creates: logs/20241023_153045-username-datasetname.log (one per job)
# Timestamp first for easy sorting with 'ls -lt | head' to see recent jobs
# Keeps all job logs for debugging, no truncation

# Option 2 (Legacy): Single log file that gets truncated upon each job
#export FCBH_DATASET_LOG_FILE="$HOME/tmp/artie/adataset.log"
# Note: This file is truncated at the start of each job (old behavior)

# Set log level (debug, info, warn, error)
export FCBH_DATASET_LOG_LEVEL="info"

# Set Python paths for MMS functionality
export FCBH_MMS_FA_PYTHON="$PWD/fcbh_env/bin/python3"
export FCBH_MMS_ASR_PYTHON="$PWD/fcbh_env/bin/python3"
export FCBH_MMS_ADAPTER_PYTHON="$PWD/fcbh_env/bin/python3"

# Set paths for various tools (we'll install these)
export FCBH_AENEAS_PYTHON="$PWD/fcbh_env/bin/python3"
export FCBH_LIBROSA_PYTHON="$PWD/fcbh_env/bin/python3"
export FCBH_WHISPER_EXE="whisper"
export FCBH_FASTTEXT_EXE="fasttext"
export FCBH_UROMAN_EXE="uroman"

# Set FFmpeg path (we'll install this)
export PATH="/opt/homebrew/bin:$PATH"

# Set Bible Brain API key (you'll need to get this)
export FCBH_DBP_KEY=""

# Set DBP MySQL connection (only needed if updating DBP timestamps)
# export DBP_MYSQL_DSN="user:password@tcp(hostname:port)/database"
export DBP_MYSQL_DSN="root:@tcp(localhost:3306)/dbp_localtest2"

echo "Environment variables set:"
echo "GOPROJ: $GOPROJ"
echo "FCBH_DATASET_DB: $FCBH_DATASET_DB"
echo "FCBH_DATASET_FILES: $FCBH_DATASET_FILES"
echo "FCBH_DATASET_TMP: $FCBH_DATASET_TMP"
echo ""
echo "Directories created at:"
echo "  Database: $FCBH_DATASET_DB"
echo "  Files: $FCBH_DATASET_FILES"
echo "  Temp: $FCBH_DATASET_TMP"
echo ""
echo "Next steps:"
echo "1. Install FFmpeg: brew install ffmpeg"
echo "2. Install Python dependencies: pip3 install -r requirements.txt"
echo "3. Get a Bible Brain API key and set FCBH_DBP_KEY"
echo "4. Test the setup with: go test ./..."


