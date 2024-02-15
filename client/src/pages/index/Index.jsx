import React from "react"
import {Navigate} from "react-router-dom"
import {HOME} from "../../config/cstModule"

function Index() {
    return <Navigate to={HOME.FULL_PATH} replace/>
}

export default Index
