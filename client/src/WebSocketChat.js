import React, { useState, useEffect } from "react";

const WebSocketChat = ({ username }) => {
    const [messages, setMessages] = useState([]);
    const [message, setMessage] = useState("");
    const [recipient, setRecipient] = useState("");
    const [ws, setWs] = useState(null);

    useEffect(() => {
        if (!username) {
            console.error("WebSocketChat received undefined username");
            return;
        }

        console.log("WebSocketChat initialized with username:", username);

        const websocket = new WebSocket("ws://localhost:8080/ws");

        websocket.onopen = () => {
            console.log("Connected to WebSocket server as", username);
            if (websocket.readyState === WebSocket.OPEN) {
                websocket.send(JSON.stringify({ username }));
            }
        };

        websocket.onmessage = (event) => {
            const receivedData = JSON.parse(event.data);
            console.log("Received WebSocket message:", receivedData);
            
            if (Array.isArray(receivedData)) {
                // If the received data is an array, it's the unread messages
                setMessages(receivedData);
            } else {
                // Otherwise, it's a new message
                setMessages((prev) => [...prev, receivedData]);
            }
        };

        websocket.onclose = () => {
            console.log("Disconnected from WebSocket server");
        };

        setWs(websocket);
        return () => websocket.close();
    }, [username]);

    const sendMessage = () => {
        if (ws && message.trim() !== "" && recipient.trim() !== "") {
            const msgData = {
                sender: username,
                recipient: recipient,
                content: message,
            };

            ws.send(JSON.stringify(msgData));
            setMessages((prev) => [...prev, msgData]); // 🔥 Show message in sender's chat
            setMessage("");
        }
    };

    return (
        <div className="p-4">
            <h2 className="text-xl font-bold">Private Chat</h2>
            <div className="border p-2 h-40 overflow-auto bg-gray-100">
                {messages.map((msg, index) => (
                    <p key={index}>
                        <strong>{msg.sender} to {msg.recipient}:</strong> {msg.content}
                    </p>
                ))}
            </div>
            <input
                type="text"
                placeholder="Recipient Username"
                value={recipient}
                onChange={(e) => setRecipient(e.target.value)}
                className="p-2 border rounded w-full mt-2"
            />
            <input
                type="text"
                placeholder="Message"
                value={message}
                onChange={(e) => setMessage(e.target.value)}
                className="p-2 border rounded w-full mt-2"
            />
            <button onClick={sendMessage} className="p-2 mt-2 bg-blue-500 text-white rounded">
                Send
            </button>
        </div>
    );
};

export default WebSocketChat;
