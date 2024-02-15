import { configureStore } from "@reduxjs/toolkit"
import rootReducer from "./reducers"
import { applyMiddleware } from "redux"
import { composeWithDevTools } from "redux-devtools-extension"
import thunk from "redux-thunk"
// import logger from "redux-logger"

export const _STORE = configureStore({
    reducer: rootReducer,
}, composeWithDevTools(applyMiddleware(thunk)))
