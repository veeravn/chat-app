import React, { useState } from "react";
import { BrowserRouter as Router, Route, Routes } from "react-router-dom";
import WebSocketChat from "./components/WebSocketChat";
import Login from "./components/Login";
import CreateUser from "./components/CreateUser";
import Navbar from "./components/Navbar";

const App = () => {
    const [user, setUser] = useState(null);

    return (
        <Router>
            <Navbar />
            <Routes>
                <Route path="/chat" element={<WebSocketChat />} />
                <Route path="/login" element={<Login setUser={setUser} />} />
                <Route path="/register" element={<CreateUser />} />
            </Routes>
        </Router>
    );
};

export default App;
