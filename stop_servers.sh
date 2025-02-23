#!/bin/bash

echo "Stopping WebSocket server..."
kill $(ps aux | grep '[g]o run server.go' | awk '{print $2}') 2>/dev/null

echo "All servers stopped."
