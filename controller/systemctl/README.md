# 1. Create or update the service file
cd $GOPROJ/controller/systemctl
sudo cp arti.service /etc/systemd/system/
sudo cp arti-failure-notify.service /etc/systemd/system/

# 2. Reload systemd to recognize the new service
sudo systemctl daemon-reload

# 3. Enable the service to start on boot
sudo systemctl enable arti.service

# 4. Start the service
sudo systemctl start arti.service

# 5. Check status
sudo systemctl status arti.service

# 6. View logs
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


