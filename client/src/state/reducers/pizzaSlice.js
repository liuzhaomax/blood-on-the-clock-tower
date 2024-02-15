import { createSlice } from "@reduxjs/toolkit"
// import { useDispatch, useSelector } from "react-redux"

const initialState = {
    toppings: ["pepper"],
    gluten: true,
}

export const pizzaSlice = createSlice({
    name: "pizza",
    initialState,
    reducers: {
        toggleGluten: state => {
            state.gluten = !state.gluten
        },
        addTopping: (state, action) => {
            state.toppings = [...state.toppings, action.payload]
        }
    },
})

export const { toggleGluten, addTopping } = pizzaSlice.actions

export default  pizzaSlice.reducer

// function App() {
// 	const pizza = useSelector(state => state.pizza)
// 	const dispatch = useDispatch()
//
// 	return (
// 		<div>
// 			<h1>Pizza</h1>
// 			{
// 				pizza.toppings.map(topping => (
// 					<div key={topping}>{topping}</div>
// 				))
// 			}
// 			<button onClick={() => dispatch(addTopping("oni"))}>Add oni</button>
// 		</div>
// 	)
// }