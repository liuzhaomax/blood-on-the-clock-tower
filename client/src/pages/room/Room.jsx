import React, {useEffect} from "react"
import "./Room.css"
import { Flex, Button } from "antd"
import {useParams} from "react-router-dom"

function Room() {
    useEffect(() => {
        setTimeout(() => {
            loadPlayerList()
        }, 100)
    })
    let { roomId } = useParams()

    const loadPlayerList = () => {
        const socket = new WebSocket("ws://localhost:8080/room/" + roomId)
        socket.onopen = function() {
            socket.send("list rooms")
        }
        socket.onmessage = function(event) {
            console.log("Received message from server:", JSON.parse(event.data))
        }
        socket.onerror = function(error) {
            console.error("WebSocket error:", error)
        }
    }

    return (
        <div id="ROOM" className="ROOM">
            <Flex className="layout" horizontal="true" gap="middle" justify="space-evenly" align="center" wrap="wrap">
                <p>我的名字：<span>{}</span></p>
            </Flex>
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
