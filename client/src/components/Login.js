import React, { useState } from "react";
import axios from "axios";

const Login = ({ setUser }) => {
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [message, setMessage] = useState("");

    const handleLogin = async () => {
        try {
            const response = await axios.post("http://localhost:8080/api/auth", {
                username,
                password,
            });

            if (response.data.success) {
                setUser(username);
                setMessage("Login successful!");
            } else {
                setMessage("Invalid credentials.");
            }
        } catch (error) {
            setMessage("Server error. Please try again.");
        }
    };

    return (
        <div className="p-4">
            <h2 className="text-xl font-bold">Login</h2>
            {message && <p className="text-red-500">{message}</p>}
            <input
                type="text"
                placeholder="Username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="p-2 border rounded w-full"
            />
            <input
                type="password"
                placeholder="Password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="p-2 border rounded w-full"
            />
            <button onClick={handleLogin} className="p-2 mt-2 bg-blue-500 text-white rounded">
                Login
            </button>
        </div>
    );
};

export default Login;
