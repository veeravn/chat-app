import React, { useState } from "react";
import { useNavigate } from "react-router-dom";

const CreateUser = () => {
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [confirmPassword, setConfirmPassword] = useState("");
    const [message, setMessage] = useState("");
    const navigate = useNavigate();

    const handleCreateUser = async () => {
        if (password !== confirmPassword) {
            setMessage("Passwords do not match");
            return;
        }

        try {
            const response = await fetch("http://localhost:8080/api/register", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ username, password }),
            });

            const data = await response.json();
            setMessage(data.message || "User created successfully");

            if (response.ok) {
                setTimeout(() => {
                    navigate("/"); // Redirect to login page
                }, 1500);
            }
        } catch (error) {
            setMessage("Error creating user");
        }
    };

    return (
        <div className="p-4 bg-white shadow-lg rounded w-80 mx-auto mt-10">
            <h2 className="text-xl font-bold mb-2">Create an Account</h2>
            {message && <p className="text-red-500">{message}</p>}
            <input
                type="text"
                placeholder="Enter Username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="p-2 border rounded w-full mt-2"
            />
            <input
                type="password"
                placeholder="Enter Password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="p-2 border rounded w-full mt-2"
            />
            <input
                type="password"
                placeholder="Confirm Password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                className="p-2 border rounded w-full mt-2"
            />
            <button onClick={handleCreateUser} className="p-2 mt-2 bg-blue-500 text-white rounded w-full">
                Register
            </button>
            <button onClick={() => navigate("/")} className="mt-2 text-blue-500 w-full">
                Back to Login
            </button>
        </div>
    );
};

export default CreateUser;
