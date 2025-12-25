#!/bin/bash -e

# Ensure HOME is set (systemd may not set it)
export HOME=/home/ec2-user

# Change to the application directory
cd /home/ec2-user/go/src/fcbh-dataset-io

# Pull latest code based on environment
if [[ "$FCBH_DATASET_QUEUE" == *"-dev" ]]; then
    git pull origin dev
else
    git pull origin main
fi

# Build and install the application
go install ./controller/queue_server

# Run the queue_server directly
exec /home/ec2-user/go/bin/queue_server