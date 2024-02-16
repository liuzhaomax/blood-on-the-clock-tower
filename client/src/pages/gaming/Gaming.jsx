import {useNavigate, useParams} from "react-router-dom"
import {Button, Flex} from "antd"
import React, {useEffect, useState} from "react"
import "./Gaming.css"

function Gaming() {
    const navigate = useNavigate()
    let { roomId } = useParams()
    const [room, setRoom] = useState(null)

    useEffect(() => {
        setTimeout(() => {
            loadRoom()
        }, 100)
    }, [room])

    const returnRoom = () => {
        navigate(`/room/${roomId}`, {
            replace: true,
            state: { roomId }
        })
    }
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

    // 已落座玩家加载
    const sit = () => {
        // 写死，只允许14名玩家
        return Array.from({length: 14}, (_, index) => {
            if (room && room.players[index]) {
                return <span key={index} className="place place-sit">{room.players[index].name}</span>
            }
        })
    }

    return (
        <div id="GAMING" className="GAMING">
            <Flex className="layout" horizontal="true" gap="middle" justify="space-evenly" align="center" wrap="wrap">
                <Button className="btn small-btn" onClick={returnRoom}>返回房间</Button>
            </Flex>
            <div>
                <p>我的名字：<span>{room === null ? "" : findPlayerName()}</span></p>
            </div>
            <p>↓ 场上玩家 ↓</p>
            <Flex className="layout" horizontal="true" gap="middle" justify="space-evenly" align="center" wrap="wrap">
                {sit()}
            </Flex>
        </div>
    )
}

export default Gaming