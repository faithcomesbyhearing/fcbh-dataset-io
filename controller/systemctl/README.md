# 1. Create the service file
sudo nano /etc/systemd/system/arti.service
# (paste in arti.service)

# 2. Make sure your script is executable
chmod +x /home/ec2-user/go/src/fcbh-dataset-io/EC2_start_script.sh

# 3. Reload systemd to recognize the new service
sudo systemctl daemon-reload

# 4. Enable the service to start on boot
sudo systemctl enable arti.service

# 5. Start the service
sudo systemctl start arti.service

# 6. Check status
sudo systemctl status arti.service

# 7. View logs
sudo journalctl -u arti.service -f

# .bash_profile must contain the following ENV's that are not to be checked into git
export OPENAI_API_KEY={open AI key}
export FCBH_DBP_KEY={Bible Brain key}
export SMTP_SENDER_EMAIL=apolyglot@fcbh.us
export SMTP_PASSWORD={email password}
export SMTP_HOST_NAME=smtp.office365.com
export SMTP_HOST_PORT=587

# must install sendemail
sudo apt install sendemail


