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
            if (room && room.state === "游戏中") {
                navigate(`/game/${roomId}`, {
                    replace: true,
                    state: `/game/${roomId}`,
                })
            }
        }, 100)
    }, [room])

    const loadRoom = () => {
        const socket = new WebSocket(`ws://192.168.1.14:8080/room/${roomId}`)
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
        const socket = new WebSocket(`ws://192.168.1.14:8080/quitRoom/${roomId}`)
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
        // 写死，只允许15名玩家
        return Array.from({length: 15}, (_, index) => {
            if (room && room.players[index]) {
                return <span key={index} className="place place-sit">{room.players[index].name}</span>
            }
            return <span key={index} className="place place-empty">空位</span>
        })
    }

    const startGame = () => {
        // if (room.players.length >= 5) {
        navigate(`/game/${roomId}`, {
            replace: true,
            state: `/game/${roomId}`,
        })
        const socket = new WebSocket(`ws://192.168.1.14:8080/startGame/${roomId}`)
        socket.onopen = function() {
            let playerInfo = {
                id: localStorage.getItem("PlayerID")
            }
            socket.send(JSON.stringify(playerInfo))
        }
        socket.onerror = function(error) {
            console.error("WebSocket error:", error)
        }
        // }
        // TODO 游戏小于5人不能开始，弹出错误提示
    }

    return (
        <div id="ROOM" className="ROOM">
            <Flex className="layout" horizontal="true" gap="middle" justify="space-evenly" align="center" wrap="wrap">
                <Button className="btn small-btn" onClick={quitRoom}>退出房间</Button>
                { room && room.host === localStorage.getItem("PlayerID")
                    ?
                    <Button className="btn small-btn start" type="primary" onClick={startGame}>开始游戏</Button>
                    :
                    <></>
                }
            </Flex>
            <p>我的名字：<span>{room === null ? "" : findPlayerName()}</span></p>
            <p>↓ 场上玩家 ↓</p>
            <div className="layout seat-wrap" align="center">
                {sit()}
            </div>
        </div>
    )
}

export default Room
