#!/bin/bash -iv

runuser --login ec2-user --shell=/bin/bash << 'EOF'
source ~/.bashrc
env
cd ~/go/src/fcbh-dataset-io
git pull
go install ./controller/queue_server
cd
nohup ~/go/bin/queue_server &
EOF
exit 0
