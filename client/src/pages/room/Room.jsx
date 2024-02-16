import React, {useEffect, useState} from "react"
import "./Room.css"
import {Button, Flex} from "antd"
import {useNavigate, useParams} from "react-router-dom"

function Room() {
    const navigate = useNavigate()
    let { roomId } = useParams()
    const [room, setRoom] = useState(null)

    useEffect(() => {
        setTimeout(() => {
            loadRoom()
        }, 100)
    }, [room])

    const loadRoom = () => {
        const socket = new WebSocket(`ws://localhost:8080/room/${roomId}`)
        socket.onopen = function() {
            socket.send("load_room")
        }
        socket.onmessage = function(event) {
            // console.log("Received message from server:", JSON.parse(event.data))
            setRoom(JSON.parse(event.data))
        }
        socket.onerror = function(error) {
            console.error("WebSocket error:", error)
        }
    }

    const findPlayerName = () => {
        if (room !== null) {
            for (let i = 0; i < room.players.length; i++) {
                if (room.players[i].id ===  localStorage.getItem("PlayerID")) {
                    return room.players[i].name
                }
            }
        }
        return ""
    }

    const quitRoom = () => {
        navigate("/home", {
            replace: true,
            state: "/home",
        })
        const socket = new WebSocket(`ws://localhost:8080/quitRoom/${roomId}`)
        socket.onopen = function() {
            let playerInfo = {
                id: localStorage.getItem("PlayerID")
            }
            socket.send(JSON.stringify(playerInfo))
        }
        socket.onerror = function(error) {
            console.error("WebSocket error:", error)
        }
    }

    const sit = () => {
        return Array.from({length: 14}, (_, index) => {
            if (room && room.players[index]) {
                return <Button key={index} className="btn place">{room.players[index].name}</Button>
            }
            return <Button key={index} className="btn place">空位</Button>
        })
    }

    return (
        <div id="ROOM" className="ROOM">
            <Flex className="layout" horizontal="true" gap="middle" justify="space-evenly" align="center" wrap="wrap">
                <Button className="btn small-btn" onClick={quitRoom}>退出房间</Button>
                <Button className="btn small-btn start" type="primary" onClick={quitRoom}>开始游戏</Button>
            </Flex>
            <p>我的名字：<span>{room === null ? "" : findPlayerName()}</span></p>
            <Flex className="layout" horizontal="true" gap="middle" justify="space-evenly" align="center" wrap="wrap">
                {sit()}
            </Flex>
        </div>
    )
}

export default Room
