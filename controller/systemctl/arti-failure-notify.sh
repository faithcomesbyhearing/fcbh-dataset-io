#!/bin/bash
# Arti service failure notification script

# Collect service status and recent logs
{
    echo "Arti service has crashed."
    echo ""
    echo "Timestamp: $(date)"
    echo "Hostname: $(hostname)"
    echo ""
    echo "=== Service Status ==="
    systemctl status arti.service
    echo ""
    echo "=== Last 100 Log Lines ==="
    journalctl -u arti.service -n 100 --no-pager
} > "$FCBH_DATASET_LOG_FILE"

# Send email with log file attached
sendemail \
    -f "$SMTP_SENDER_EMAIL" \
    -t "$FAILURE_RECIPIENTS" \
    -u "ALERT: Arti Service Crashed" \
    -m "Arti service has failed and been restarted. See attached log file for details." \
    -a "FCBH_DATASET_LOG_FILE" \
    -s "$SMTP_HOST_NAME:$SMTP_HOST_PORT" \
    -o tls=yes \
    -xu "$SMTP_SENDER_EMAIL" \
    -xp "$SMTP_PASSWORD"

#sleep 10
#systemctl start arti.service

exit 0

