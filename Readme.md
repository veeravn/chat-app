# **📌 WebSocket Chat Server**
A **scalable WebSocket chat application** with:
- **Real-time messaging**
- **Multi-replica WebSocket servers (Docker)**
- **Message persistence in Cassandra**
- **User connection tracking in Redis**
- **WebSocket message forwarding**

---

## **🚀 Features**
✅ **Real-time messaging** with WebSockets  
✅ **Message persistence** in Cassandra  
✅ **Multi-replica WebSocket servers** using Docker  
✅ **User presence tracking** with Redis  
✅ **Message forwarding** between WebSocket servers  
✅ **Automatic message marking as read**  

---

## **🛠 Project Structure**
```
/websocket-server
│── main.go                  # Entry point
│── config/
│   ├── config.go             # Loads environment variables & config
│── database/
│   ├── cassandra.go          # Cassandra database functions
│   ├── redis.go              # Redis functions
│── handlers/
│   ├── websocket.go          # WebSocket connection handlers
│   ├── messages.go           # Message processing functions
│── models/
│   ├── message.go            # Message struct
│── utils/
│   ├── logger.go             # Logging functions
│── go.mod                    # Go module file
│── go.sum                    # Go dependencies
│── Dockerfile                 # Docker build file
│── docker-compose.yml         # Docker Compose for multi-replica setup
│── README.md                 # Project documentation
```

---

## **📦 Prerequisites**
🔹 **Go 1.18+**  
🔹 **Docker & Docker Compose**  
🔹 **Cassandra Database**  
🔹 **Redis**  

---

## **⚙️ Installation**
### **1️⃣ Clone the Repository**
```sh
git clone https://github.com/your-repo/websocket-chat.git
cd websocket-chat
```

### **2️⃣ Set Up Environment Variables**
Create a **`.env`** file with the following:
```env
REDIS_HOST=redis
CASSANDRA_HOST=cassandra
WEBSOCKET_SERVER=ws://localhost:8080
```

---

## **🚀 Run the WebSocket Server**
### **1️⃣ Run With Docker Compose**
```sh
docker-compose up --scale ws-server=3 --build
```
This will:
- Start **3 WebSocket server replicas**
- Start **Cassandra & Redis**
- Load balance WebSocket traffic

### **2️⃣ Run Without Docker (Local)**
```sh
go run main.go
```
This will start the **WebSocket server on port 8080**.

---

## **📝 Usage**
### **1️⃣ Connect to WebSocket**
```sh
wscat -c ws://localhost:8080/ws
```
or use a WebSocket client in the browser.

### **2️⃣ Send a Message**
Send a **JSON message**:
```json
{
    "sender": "user1",
    "recipient": "user2",
    "content": "Hello from user1!"
}
```

### **3️⃣ Check Messages in Cassandra**
```sh
docker exec -it cassandra cqlsh
SELECT * FROM chat.messages;
```

---

## **📡 WebSocket API**
### **1️⃣ WebSocket Connection**
- **Endpoint:** `ws://localhost:8080/ws`
- **Message Format (JSON)**
```json
{
    "sender": "user1",
    "recipient": "user2",
    "content": "Hello!"
}
```

### **2️⃣ WebSocket Events**
✅ **Message Sent** → Message is forwarded or stored  
✅ **Message Received** → Marked as read in Cassandra  
✅ **User Connected** → Stored in Redis  

---

## **🔍 Debugging**
### **1️⃣ Check Redis User Mapping**
```sh
docker exec -it redis redis-cli
keys user:*
get user:user1
```

### **2️⃣ Check WebSocket Logs**
```sh
docker logs websocket-server_1
```

---

## **🔄 Scaling**
To scale WebSocket servers:
```sh
docker-compose up --scale ws-server=5 --build
```
This launches **5 WebSocket replicas**, ensuring **high availability**.
