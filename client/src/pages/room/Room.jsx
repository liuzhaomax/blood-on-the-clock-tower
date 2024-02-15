import React, {useEffect, useState} from "react"
import "./Room.css"
import { Flex, Button } from "antd"
import {useNavigate, useParams} from "react-router-dom"

function Room() {
    const navigate = useNavigate()
    let { roomId } = useParams()
    const [room, setRoom] = useState(null)

    useEffect(() => {
        setTimeout(() => {
            loadPlayerList()
        }, 500)
    }, [room])

    const loadPlayerList = () => {
        const socket = new WebSocket("ws://localhost:8080/room/" + roomId)
        socket.onopen = function() {
            socket.send("list rooms")
        }
        socket.onmessage = function(event) {
            // console.log("Received message from server:", JSON.parse(event.data))
            setRoom(JSON.parse(event.data))
        }
        socket.onerror = function(error) {
            console.error("WebSocket error:", error)
        }
    }

    const quitRoom = () => {
        navigate("/home", {
            replace: true,
            state: "/home",
        })
    }

    return (
        <div id="ROOM" className="ROOM">
            <Flex className="layout" horizontal="true" gap="middle" justify="space-evenly" align="center" wrap="wrap">
                <Button className="btn small-btn" onClick={quitRoom}>退出房间</Button>
                <Button className="btn small-btn" type="primary" onClick={quitRoom}>开始游戏</Button>
            </Flex>
            <p>我的名字：<span>{room === null ? "" : room.players[0].name}</span></p>
            <Flex className="layout" horizontal="true" gap="middle" justify="space-evenly" align="center" wrap="wrap">
                <Button className="btn" type="dashed">abc</Button>
            </Flex>
            <Flex className="layout" horizontal="true" gap="middle" justify="space-evenly" align="center" wrap="wrap">
                <Button className="btn" type="dashed">abc</Button>
            </Flex>
        </div>
    )
}

export default Room
