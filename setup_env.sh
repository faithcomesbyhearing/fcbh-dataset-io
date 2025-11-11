#!/bin/bash

# sample FCBH Dataset I/O Environment Setup

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
# Set log level (debug, info, warn, error)
export FCBH_DATASET_LOG_LEVEL="info"
# Single log file that gets truncated upon each job
#export FCBH_DATASET_LOG_FILE="$HOME/tmp/artie/adataset.log"
# OR: timestamped logs in a directory (no truncation, overules above if set)
export FCBH_DATASET_LOG_DIR="$HOME/tmp/artie/logs"

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

# Set Bible Brain API key (you'll need to get this)
export FCBH_DBP_KEY=""

# Set DBP MySQL connection (only needed for dbp_update)
# export DBP_MYSQL_DSN="user:password@tcp(hostname:port)/database"

