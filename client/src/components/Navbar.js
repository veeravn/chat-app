import React from "react";
import { Link } from "react-router-dom";

const Navbar = () => {
    return (
        <nav className="p-4 bg-gray-800 text-white flex justify-between">
            <h1 className="text-xl">WebSocket Chat</h1>
            <div>
                <Link to="/chat" className="mr-4">Chat</Link>
                <Link to="/login" className="mr-4">Login</Link>
                <Link to="/register">Register</Link>
            </div>
        </nav>
    );
};

export default Navbar;
