import React, { useState } from "react";
import axios from "axios";

const CreateUser = () => {
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [confirmPassword, setConfirmPassword] = useState("");
    const [message, setMessage] = useState("");

    const handleCreateUser = async () => {
        if (!username || !password) {
            setMessage("Username and password are required.");
            return;
        }

        if (password !== confirmPassword) {
            setMessage("Passwords do not match.");
            return;
        }

        try {
            const response = await axios.post("http://load-balancer:8080/api/register", {
                username,
                password,
            });

            if (response.data.success) {
                setMessage("User created successfully! You can now log in.");
            } else {
                setMessage(response.data.message || "Error creating user.");
            }
        } catch (error) {
            setMessage("Server error. Please try again.");
        }
    };

    return (
        <div className="flex flex-col items-center p-4">
            <h2 className="text-2xl font-bold">Create New User</h2>
            {message && <p className="text-red-500">{message}</p>}
            <input
                type="text"
                placeholder="Username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="p-2 border rounded mt-2"
            />
            <input
                type="password"
                placeholder="Password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="p-2 border rounded mt-2"
            />
            <input
                type="password"
                placeholder="Confirm Password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                className="p-2 border rounded mt-2"
            />
            <button onClick={handleCreateUser} className="px-4 py-2 bg-blue-600 text-white rounded mt-2">
                Create Account
            </button>
        </div>
    );
};

export default CreateUser;
