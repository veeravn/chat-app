import React, { useState } from "react";
import { BrowserRouter as Router, Route, Routes } from "react-router-dom";
import WebSocketChat from "./WebSocketChat";
import Login from "./Login";
import Navbar from "./Navbar";

const App = () => {
    const [username, setUsername] = useState("");

    const handleLogin = (user) => {
        if (user && user.trim() !== "") {
            setUsername(user.trim()); // Set and trim username
            console.log("User logged in:", user);
        }
    };

    const handleLogout = () => {
        setUsername(""); // Clear username on logout
        console.log("User logged out");
    };

    return (
        <div>
            <Navbar username={username} onLogout={handleLogout} />
            <div className="p-4">
                {!username ? (
                    <Login onLogin={handleLogin} />
                ) : (
                    <WebSocketChat username={username} />
                )}
            </div>
        </div>
    );
};

export default App;
