FROM golang:1.19

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN chmod +x start_servers.sh
RUN chmod +x log_servers.sh

EXPOSE 8080

CMD ["/bin/bash", "-c", "./start_servers.sh $PORT"]
