import React from "react"
import "./App.css"
import {BrowserRouter as Router} from "react-router-dom"
import routes from "./routes"

function App() {
    return (
        <div id="App" className="App">
            <Router>{ routes }</Router>
        </div>
    )
}

export default App
