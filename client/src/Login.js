import React, { useState } from "react";
import { useNavigate } from "react-router-dom";

const Login = ({ onLogin }) => {
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [error, setError] = useState("");
    const navigate = useNavigate();

    const handleLogin = async () => {
        if (username.trim() === "" || password.trim() === "") {
            setError("Username and password are required");
            return;
        }

        try {
            const response = await fetch("http://localhost:8080/api/auth", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ username, password }),
            });

            if (!response.ok) {
                throw new Error("Invalid username or password");
            }

            const data = await response.json();
            console.log("Login successful:", data);
            onLogin(username);
        } catch (err) {
            console.error("Login error:", err);
            setError(err.message);
        }
    };

    return (
        <div className="p-4 bg-white shadow-lg rounded w-80 mx-auto mt-10">
            <h2 className="text-xl font-bold">Login</h2>
            {error && <p className="text-red-500">{error}</p>}
            <input
                type="text"
                placeholder="Enter your username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="p-2 border rounded w-full mt-2"
            />
            <input
                type="password"
                placeholder="Enter your password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="p-2 border rounded w-full mt-2"
            />
            <button onClick={handleLogin} className="p-2 mt-2 bg-blue-500 text-white rounded w-full">
                Login
            </button>
            <button onClick={() => navigate("/create-user")} className="mt-2 text-blue-500 w-full">
                Create an Account
            </button>
        </div>
    );
};

export default Login;
