#!/bin/bash
# Arti service failure notification script

RECIPIENTS="gary@shortsands.com,jrstear@fcbhmail.org"
SUBJECT="ALERT: Arti Service Failed and Restarted"
LOGFILE="/home/ec2-user/dataset.log"

# Collect service status and recent logs
{
    echo "Arti service has failed and been restarted."
    echo ""
    echo "Timestamp: $(date)"
    echo "Hostname: $(hostname)"
    echo ""
    echo "=== Service Status ==="
    systemctl status arti.service
    echo ""
    echo "=== Last 1000 Log Lines ==="
    journalctl -u arti.service -n 1000 --no-pager
} > "$LOGFILE"

# Send email with log file attached
#mail -s "$SUBJECT" -a "$LOGFILE" "$RECIPIENTS" < "$LOGFILE"
sendemail \
    -f "$SMTP_SENDER_EMAIL" \
    -t "$RECIPIENTS" \
    -u "$SUBJECT" \
    -m "Arti service has failed and been restarted. See attached log file for details." \
    -a "$LOGFILE" \
    -s "$SMTP_HOST_NAME:$SMTP_HOST_PORT" \
    -o tls=yes \
    -xu "$SMTP_SENDER_EMAIL" \
    -xp "$SMTP_PASSWORD"

exit 0

