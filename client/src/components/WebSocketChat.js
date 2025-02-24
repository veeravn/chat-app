import React, { useState, useEffect } from "react";

const WebSocketChat = ({ username }) => {
    const [messages, setMessages] = useState([]);
    const [message, setMessage] = useState("");
    const [recipient, setRecipient] = useState("");
    const [typing, setTyping] = useState(false);
    const [isTyping, setIsTyping] = useState(false);
    const [socket, setSocket] = useState(null);

    useEffect(() => {
        const ws = new WebSocket("ws://localhost:8080/ws");

        ws.onopen = () => {
            console.log("Connected to WebSocket server");
            ws.send(JSON.stringify({ sender: username }));
        };

        ws.onmessage = (event) => {
            const receivedMsg = JSON.parse(event.data);
            if (receivedMsg.type === "typing") {
                setIsTyping(receivedMsg.sender !== username);
                return;
            }
            setMessages((prev) => [...prev, receivedMsg]);
        };

        setSocket(ws);

        return () => ws.close();
    }, [username]);

    const sendMessage = () => {
        if (socket && message.trim() !== "" && recipient.trim() !== "") {
            const msgData = {
                sender: username,
                recipient: recipient,
                content: message,
                read: false
            };
            socket.send(JSON.stringify(msgData));
            setMessage("");
        }
    };

    const handleTyping = () => {
        if (socket && !typing) {
            setTyping(true);
            socket.send(JSON.stringify({ type: "typing", sender: username }));
            setTimeout(() => setTyping(false), 2000);
        }
    };

    return (
        <div className="p-4">
            <h2 className="text-xl font-bold">Private Chat</h2>
            <div className="border p-2 h-40 overflow-auto bg-gray-100">
                {messages.map((msg, index) => (
                    <p key={index} className={msg.read ? "text-gray-500" : "text-black"}>
                        <strong>{msg.sender} to {msg.recipient}:</strong> {msg.content} {msg.read && "✔️"}
                    </p>
                ))}
                {isTyping && <p className="text-gray-500">{recipient} is typing...</p>}
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
                onKeyDown={handleTyping}
                className="p-2 border rounded w-full mt-2"
            />
            <button onClick={sendMessage} className="p-2 mt-2 bg-blue-500 text-white rounded">
                Send
            </button>
        </div>
    );
};

export default WebSocketChat;
