# systemctl is the preferred way to start, stop and restart arti.
# This readme describes how to install and use systemctl.

# When installed, the following commands will be available:
# sudo systemctl start arti.service
# sudo systemctl stop arti.service
# sudo systemctl restart arti.service

# When these instructions are complete the current startup script needs to be modified
# to remove the startup of arti. 
cat /etc/rc.local

# must install sendemail
sudo apt install sendemail

# 1. Create or update the service file
cd /home/ec2-user/go/src/fcbh-dataset-io/controller/systemctl
sudo cp arti.service /etc/systemd/system/
sudo cp arti-failure-notify.service /etc/systemd/system/

# 2. Reload systemd to recognize the new service
# And do this anytime either service file is updated
sudo systemctl daemon-reload

# 3. Enable the service to start on boot
sudo systemctl enable arti.service

# 4. Start the service
sudo systemctl start arti.service

# 5. Check status
sudo systemctl status arti.service

# 6. View logs
sudo journalctl -u arti.service -f
sudo journalctl -u arti-failure-notify.service

# /etc/arti.env must contain the following variables
sudo mkdir /etc/arti
sudo vi /etc/arti/arti.env
i
OPENAI_API_KEY={open AI key}
FCBH_DBP_KEY={Bible Brain key}
FAILURE_RECIPIENTS={email addresses comma separated}
SMTP_SENDER_EMAIL=apolyglot@fcbh.us
SMTP_PASSWORD={email password}
SMTP_HOST_NAME=smtp.office365.com
SMTP_HOST_PORT=587
FCBH_DATASET_LOG_FILE=/home/ec2-user/dataset.log
FCBH_DATASET_QUEUE={queue s3 bucket name}
FCBH_DATASET_IO_BUCKET={output s3 bucket name}
FCBH_SQS_URL_PREFIX=https://sqs.us-west-2.amazonaws.com/078432969830/ # example
escZZ




