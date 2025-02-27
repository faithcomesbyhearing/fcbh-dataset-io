#!/bin/bash -v
cd /home/ec2-user/go/src/fcbh-dataset-io
git pull
/usr/local/go/bin/go install ./controller/queue_server
cd /home/ec2-user
nohup /home/ec2-user/go/bin/queue_server &