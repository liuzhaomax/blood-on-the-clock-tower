import React from "react"
// import {Suspense} from "react"
import { Routes, Route } from "react-router-dom"
import Index from "../pages/index/Index"
import NotFound from "../pages/notFound/NotFound"
import Home from "../pages/home/Home"
import HomeLayout from "../layouts/HomeLayout"
import { HOME } from "../config/cstModule"
import Room from "../pages/room/Room"
import Gaming from "../pages/gaming/Gaming"

// const lazyLoad = path => {
//     const Comp = React.lazy(() => import(`../${path}`))
//     return (
//         <Suspense fallback={<>加载中...</>}>
//             <Comp />
//         </Suspense>
//     )
// }

export default (
    <Routes>
        <Route path="/" element={<Index/>}/>
        <Route element={<HomeLayout/>}>
            <Route path={HOME.PATH} element={<Home/>}/>
            <Route path="/room/:roomId" element={<Room/>}/>
            <Route path="/game/:roomId" element={<Gaming/>}/>
        </Route>
        <Route path="*" element={<NotFound/>}/>
    </Routes>
)