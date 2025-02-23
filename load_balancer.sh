#!/bin/bash

PORT=${1:-8080}  # Default port is 8080 if not specified

echo "Starting WebSocket load balancer on port $PORT..."
go run load_balancer.go --port=$PORT &

wait
