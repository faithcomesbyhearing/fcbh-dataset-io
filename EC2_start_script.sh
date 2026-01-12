#!/bin/bash -v

runuser --login ec2-user --shell=/bin/bash << 'EOF'
env
cd ~/go/src/fcbh-dataset-io
if [[ "$FCBH_DATASET_QUEUE" == *"-dev" ]]; then
    git pull origin dev
else
    git fetch --tags
    git checkout v1.1.12
#    git pull origin main
fi
go install ./controller/queue_server
cd
nohup ~/go/bin/queue_server &
EOF
exit 0

