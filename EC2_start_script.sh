#!/bin/bash -v

su - ec2-user << 'EOF'
source ~/.bashrc
env
cd ~/go/src/fcbh-dataset-io
git pull
go install ./controller/queue_server
cd
nohup ~go/bin/queue_server &
EOF
exit 0
