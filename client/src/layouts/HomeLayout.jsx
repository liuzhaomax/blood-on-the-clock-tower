import React from "react"
import "./HomeLayout.css"
import { Outlet } from "react-router-dom"

function HomeLayout() {
    return (
        <div className="HomeLayout">
            <Outlet />
        </div>
    )
}

export default HomeLayout

