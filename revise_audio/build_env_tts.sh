#!/bin/bash
# Build environment for MMS TTS (Text-to-Speech)
# Reuses the same environment as MMS ASR

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MMS_ASR_ENV="${SCRIPT_DIR}/../mms/mms_asr/build_env.sh"

if [ -f "$MMS_ASR_ENV" ]; then
    echo "Using MMS ASR environment setup..."
    source "$MMS_ASR_ENV"
else
    echo "Warning: MMS ASR build_env.sh not found, using defaults"
    export FCBH_MMS_ASR_PYTHON="python3"
fi

# Additional dependencies for TTS
echo "Checking TTS dependencies..."
python3 -c "import scipy" 2>/dev/null || {
    echo "Installing scipy..."
    pip install scipy
}

echo "✅ MMS TTS environment ready"
echo "   Python: ${FCBH_MMS_ASR_PYTHON:-python3}"

