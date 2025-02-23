#!/bin/bash

echo "Stopping WebSocket servers..."
kill $(ps aux | grep '[g]o run websocket_server_with_health.go' | awk '{print $2}') 2>/dev/null

echo "Stopping Load Balancer..."
kill $(ps aux | grep '[g]o run load_balancer.go' | awk '{print $2}') 2>/dev/null

echo "All WebSocket servers and load balancer stopped."

