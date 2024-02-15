import React, { useEffect, useState } from "react"
import NoFound from "antd/es/result/noFound"
import {Button} from "antd"
import { useNavigate } from "react-router-dom"

function NotFound() {
    const [secondsToGo, setSecondsToGo] = useState(5)
    const [timer, setTimer] = useState(null)
    const [timerOut, setTimerOut] = useState(null)

    const navigate = useNavigate()

    useEffect(() => {
        setTimer(
            setInterval(() => {
                setSecondsToGo(secondsToGo - 1)
            }, 1000)
        )
        setTimerOut(
            setTimeout(() => {
                navigate(-1)
            }, secondsToGo * 1000)
        )
        return () => {
            clearTimeout(timerOut)
            clearInterval(timer)
        }
    },[secondsToGo])

    const onClickGoBack = () => {
        navigate(-1)
    }

    return (
        <div className="NotFound">
            <h3>寄！404</h3>
            <NoFound/>
            <p>将于 <span id="nf-cound-down">{secondsToGo}</span> 秒后跳转到之前的页面</p>
            <Button type="primary" onClick={onClickGoBack}>返回</Button>
        </div>
    )
}

export default NotFound
