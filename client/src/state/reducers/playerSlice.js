import { createSlice } from "@reduxjs/toolkit"

const initialState = {
    id: "",
    name: "",
}

export const playerSlice = createSlice({
    name: "player",
    initialState,
    reducers: {
        signPlayerName: (state, action) => {
            state.name = action.payload
        },
        signPlayerId: (state, action) => {
            state.id = action.payload
        }
    },
})

export const { signPlayerId, signPlayerName } = playerSlice.actions

export default  playerSlice.reducer