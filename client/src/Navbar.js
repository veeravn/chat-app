import React from "react";
import CreateUser from "./CreateUser";
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
                    <button onClick={() => setShowCreateUser(!showCreateUser)} className="bg-green-500 px-3 py-1 rounded">
                        {showCreateUser ? "Close" : "Create User"}
                    </button>
                )}
            </div>
            {showCreateUser && <CreateUser />}
        </nav>
    );
};

export default Navbar;
