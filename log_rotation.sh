#!/bin/bash

LOG_DIR="logs"
MAX_LOG_SIZE=10485760  # 10MB
BACKUP_COUNT=5

mkdir -p $LOG_DIR

rotate_log() {
    local log_file=$1
    if [ -f "$log_file" ] && [ $(stat -c%s "$log_file") -ge $MAX_LOG_SIZE ]; then
        for ((i=BACKUP_COUNT-1; i>0; i--)); do
            if [ -f "$log_file.$i" ]; then
                mv "$log_file.$i" "$log_file.$((i+1))"
            fi
        done
        mv "$log_file" "$log_file.1"
        echo "Log rotated: $log_file"
    fi
}

while true; do
    for log_file in $LOG_DIR/*.log; do
        rotate_log "$log_file"
    done
    sleep 60  # Check logs every minute
done
