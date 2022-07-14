import './App.css';
import './audio/AudioPlayer'
import { AudioPage } from "./audio/AudioPage";
import { Route, Routes, Link } from "react-router-dom"
import React from 'react';
import { ResumePage } from "./resume/ResumePage";
import { HomePage } from "./home/HomePage";
import { AdminPage } from "./user/AdminPage";
import { KeyOfDay } from "./keyOfDay/KeyOfDay";
import { NotFoundPage } from "./error/NotFoundPage";

function App() {
	return (
		<div className="center">
			<div className="container">
				<div className="App">
					<ul className="navbar">
						<li>
							<Link to="/">Home</Link>
						</li>
						<li>
							<Link to="/music">Music</Link>
						</li>
						<li>
							<Link to="/resume">CV</Link>
						</li>
					</ul>
					<Routes>
						<Route path="/" element={<HomePage />} />
						<Route path="/music" element={<AudioPage />} />
						<Route path="/resume" element={<ResumePage />} />
						<Route path="/admin" element={<AdminPage />} />
						<Route path="/kod" element={<KeyOfDay />} />
						<Route path="/*" element={<NotFoundPage />} />
					</Routes>
				</div>
			</div>
		</div>
	);
}

export default App;
