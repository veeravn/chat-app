import React from "react";

const Navbar = ({ username, onLogout }) => {
    return (
        <nav className="bg-gray-800 text-white p-4 flex justify-between">
            <h1 className="text-lg font-bold">Chat App</h1>
            {username && (
                <div className="flex items-center">
                    <span className="mr-4">Logged in as: {username}</span>
                    <button onClick={onLogout} className="bg-red-500 px-3 py-1 rounded">
                        Logout
                    </button>
                </div>
            )}
        </nav>
    );
};

export default Navbar;
