import { combineReducers } from "redux"
import pizzaSlice from "./reducers/pizzaSlice"
import playerSlice from "./reducers/playerSlice"

const rootReducer = combineReducers({
    pizzaSlice,
    playerSlice,
})

export default rootReducer