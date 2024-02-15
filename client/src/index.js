import React from "react"
import { createRoot } from "react-dom/client"
import "./index.css"
import App from "./app/App"
import { Provider } from "react-redux"
import env from "./env/env.json"
import axios from "axios"
import config from "./config/config"
import { _STORE } from "./state/store"

// https upgrading config
if (env.run_mode === "prod") {
    const head = document.getElementsByTagName("head")[0]
    const https = document.createElement("meta")
    https.setAttribute("http-equiv", "Content-Security-Policy")
    https.setAttribute("content", "upgrade-insecure-requests")
    head.appendChild(https)
}

axios.defaults.baseURL = config.beBaseUrl

createRoot(document.getElementById("root")).render(
    <Provider store={_STORE}>
        <App store={_STORE} />
    </Provider>
)