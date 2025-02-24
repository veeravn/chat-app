import React, { useState } from "react";
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import CreateUser from "./CreateUser";
import Login from "./Login";
import Navbar from "./Navbar";
import WebSocketChat from "./WebSocketChat";
import { useNavigate } from "react-router-dom";

const AppWrapper = () => {
    return (
        <Router>
            <App />
        </Router>
    );
};

const App = () => {
    const [username, setUsername] = useState("");
    const navigate = useNavigate();

    const handleLogin = (user) => {
        if (user && user.trim() !== "") {
            setUsername(user.trim());
            console.log("User logged in:", user);
            navigate("/chat"); // Redirect to chat page after login
        }
    };

    const handleLogout = () => {
        setUsername("");
        console.log("User logged out");
        navigate("/"); // Redirect to login page after logout
    };

    return (
        <>
            <Navbar username={username} onLogout={handleLogout} />
            <div className="p-4">
                <Routes>
                    <Route path="/create-user" element={<CreateUser />} />
                    <Route path="/chat" element={username ? <WebSocketChat username={username} /> : <Login onLogin={handleLogin} />} />
                    <Route path="/" element={!username ? <Login onLogin={handleLogin} /> : <p>Welcome, {username}!</p>} />
                </Routes>
            </div>
        </>
    );
};

export default AppWrapper;
