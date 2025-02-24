import React from "react";
import { Link } from "react-router-dom";

const Navbar = ({ username, onLogout }) => {
    return (
        <nav className="bg-gray-800 text-white p-4 flex justify-between">
            <h1 className="text-lg font-bold">Chat App</h1>
            <div className="flex items-center">
                {username ? (
                    <>
                        <span className="mr-4">Logged in as: {username}</span>
                        <button onClick={onLogout} className="bg-red-500 px-3 py-1 rounded mr-2">
                            Logout
                        </button>
                    </>
                ) : (
                    <Link to="/create-user" className="bg-green-500 px-3 py-1 rounded">
                        Create Account
                    </Link>
                )}
            </div>
        </nav>
    );
};

export default Navbar;
