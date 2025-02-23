#!/bin/bash

LOG_DIR="logs"
mkdir -p $LOG_DIR

# Number of WebSocket server instances
NUM_SERVERS=3

# Start WebSocket servers with logging
for ((i=1; i<=NUM_SERVERS; i++)); do
    PORT=$((8080 + i))
    echo "Starting WebSocket server on port $PORT..."
    go run websocket_server_with_health.go --port=$PORT > "$LOG_DIR/server_$PORT.log" 2>&1 &
done

echo "Starting Load Balancer on port 8080..."
go run load_balancer.go > "$LOG_DIR/load_balancer.log" 2>&1 &

echo "All servers started with logs stored in $LOG_DIR/"
wait

