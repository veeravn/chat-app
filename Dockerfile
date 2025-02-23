FROM golang:1.19

# Set the working directory
WORKDIR /app

# Copy the Go modules files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the WebSocket server
RUN go build -o websocket_server websocket_server_with_health.go

# Expose ports for WebSocket servers
EXPOSE 8081 8082 8083 8080

# Start multiple WebSocket servers and the load balancer with health checks
CMD ["/bin/bash", "-c", "./log_servers.sh & ./start_servers.sh"]

