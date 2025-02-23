#!/bin/bash

LOG_DIR="logs"
mkdir -p $LOG_DIR

echo "Starting WebSocket server with logging..."
go run server.go > "$LOG_DIR/websocket.log" 2>&1 &

wait
