import React, { useState, useEffect, useRef } from "react";
import axios from "axios";
import "tailwindcss/tailwind.css";

const WebSocketChat = () => {
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [recipient, setRecipient] = useState("");
    const [message, setMessage] = useState("");
    const [messages, setMessages] = useState([]);
    const [typing, setTyping] = useState("");
    const [onlineUsers, setOnlineUsers] = useState([]);
    const [authenticated, setAuthenticated] = useState(false);
    const [theme, setTheme] = useState("light");
    const socketRef = useRef(null);

    useEffect(() => {
        document.body.className = theme === "dark" ? "bg-gray-900 text-white" : "bg-white text-black";
    }, [theme]);

    const authenticateUser = async () => {
        if (!username || !password) {
            alert("Please enter a username and password");
            return;
        }
        try {
            const response = await axios.post("http://localhost:5000/api/auth", { username, password });
            if (response.data.success) {
                setAuthenticated(true);
                connectWebSocket();
            } else {
                alert("Invalid credentials");
            }
        } catch (error) {
            alert("Authentication failed");
        }
    };

    const connectWebSocket = () => {
        if (!authenticated) {
            alert("Please log in first");
            return;
        }
        socketRef.current = new WebSocket("ws://localhost:8080/ws");
        socketRef.current.onopen = () => {
            console.log("Connected to WebSocket server");
            socketRef.current.send(JSON.stringify({ type: "join", user: username }));
        };
        socketRef.current.onmessage = (event) => {
            const data = JSON.parse(event.data);
            if (data.type === "message" && data.recipient === username) {
                setMessages((prevMessages) => [...prevMessages, `${data.user} -> You: ${data.message} (Delivered)`]);
                socketRef.current.send(JSON.stringify({ type: "read_receipt", user: username, sender: data.user }));
            } else if (data.type === "typing") {
                setTyping(`${data.user} is typing...`);
                setTimeout(() => setTyping(""), 3000);
            } else if (data.type === "online_users") {
                setOnlineUsers(data.users);
            } else if (data.type === "read_receipt" && data.user === username) {
                alert(`Your message was read by ${data.sender}`);
            }
        };
        socketRef.current.onclose = () => {
            console.log("Disconnected from WebSocket server");
        };
    };

    const sendMessage = () => {
        if (!recipient) {
            alert("Please enter a recipient username");
            return;
        }
        if (message && socketRef.current) {
            socketRef.current.send(JSON.stringify({ type: "message", user: username, message, recipient }));
            setMessage("");
        }
    };

    const sendTypingStatus = () => {
        if (socketRef.current) {
            socketRef.current.send(JSON.stringify({ type: "typing", user: username }));
        }
    };

    return (
        <div className="flex flex-col items-center p-4">
            <h2 className="text-2xl font-bold">WebSocket Private Chat</h2>
            <button className="mb-4 px-4 py-2 bg-gray-600 text-white rounded" onClick={() => setTheme(theme === "light" ? "dark" : "light")}>
                Toggle {theme === "light" ? "Dark" : "Light"} Mode
            </button>
            {!authenticated ? (
                <div className="flex flex-col gap-2">
                    <input className="p-2 border rounded" type="text" placeholder="Enter your username" value={username} onChange={(e) => setUsername(e.target.value)} />
                    <input className="p-2 border rounded" type="password" placeholder="Enter your password" value={password} onChange={(e) => setPassword(e.target.value)} />
                    <button className="px-4 py-2 bg-blue-600 text-white rounded" onClick={authenticateUser}>Login</button>
                </div>
            ) : (
                <div className="flex flex-col gap-2">
                    <p className="text-sm italic">{typing}</p>
                    <p className="text-sm">Online Users: {onlineUsers.join(", ")}</p>
                    <ul className="border p-2 h-40 overflow-y-auto">
                        {messages.map((msg, index) => (
                            <li key={index} className="p-1 border-b">{msg}</li>
                        ))}
                    </ul>
                    <input className="p-2 border rounded" type="text" placeholder="Recipient username" value={recipient} onChange={(e) => setRecipient(e.target.value)} />
                    <input className="p-2 border rounded" type="text" placeholder="Type a message..." value={message} onChange={(e) => setMessage(e.target.value)} onInput={sendTypingStatus} />
                    <button className="px-4 py-2 bg-green-600 text-white rounded" onClick={sendMessage}>Send</button>
                </div>
            )}
        </div>
    );
};

export default WebSocketChat;