#!/bin/bash

PORT=${1:-8080}  # Default port is 8080 if not specified

echo "Starting WebSocket server with authentication on port $PORT..."
go run server.go --port=$PORT &

wait
