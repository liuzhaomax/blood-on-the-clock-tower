import React from "react"
import {Suspense} from "react"
import { Routes, Route } from "react-router-dom"
import Index from "../pages/index/Index"
import NotFound from "../pages/notFound/NotFound"
import Home from "../pages/home/Home"
import HomeLayout from "../layouts/HomeLayout"
import { HOME } from "../config/cstModule"
// import Room from "../pages/room/Room"
// import Gaming from "../pages/gaming/Gaming"
// import Review from "../pages/review/Review"

const lazyLoad = path => {
    const Comp = React.lazy(() => import(`../${path}`))
    return (
        <Suspense fallback={<>加载中...</>}>
            <Comp />
        </Suspense>
    )
}

export default (
    <Routes>
        <Route path="/" element={<Index/>}/>
        <Route element={<HomeLayout/>}>
            <Route path={HOME.PATH} element={<Home/>}/>
            <Route path="/room/:roomId" element={lazyLoad("pages/room/Room")}/>
            <Route path="/game/:roomId" element={lazyLoad("pages/gaming/Gaming")}/>
            <Route path="/review/:roomId" element={lazyLoad("pages/review/Review")}/>
        </Route>
        <Route path="*" element={<NotFound/>}/>
    </Routes>
)