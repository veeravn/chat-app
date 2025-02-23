#!/bin/bash

# Number of WebSocket server instances
NUM_SERVERS=3

# Starting WebSocket servers on different ports
for ((i=1; i<=NUM_SERVERS; i++)); do
    PORT=$((8080 + i))
    echo "Starting WebSocket server on port $PORT..."
    go run websocket_server_with_health.go --port=$PORT &
done

echo "Starting Load Balancer on port 8080..."
go run load_balancer.go &

echo "All WebSocket servers started with logging."
wait

