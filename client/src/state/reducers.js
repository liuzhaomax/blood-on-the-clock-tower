import { combineReducers } from "redux"
import pizzaSlice from "./reducers/pizzaSlice"

const rootReducer = combineReducers({
    pizzaSlice,
})

export default rootReducer