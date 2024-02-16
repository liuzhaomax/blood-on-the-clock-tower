import {useNavigate, useParams} from "react-router-dom"
import {Button} from "antd"
import React, {useEffect, useState} from "react"
import "./Gaming.css"

function Gaming() {
    const navigate = useNavigate()
    let { roomId } = useParams()
    const [game, setGame] = useState(null)

    useEffect(() => {
        loadGame()
    }, [])

    const returnRoom = () => {
        navigate(`/room/${roomId}`, {
            replace: true,
            state: { roomId }
        })
    }

    const loadGame = () => {
        const socket = new WebSocket(`ws://localhost:8080/game/${roomId}`)
        socket.onopen = function() {
            socket.send("load_game")
        }
        socket.onmessage = function(event) {
            // console.log("Received message from server:", JSON.parse(event.data))
            setGame(JSON.parse(event.data))
        }
        socket.onerror = function(error) {
            console.error("WebSocket error:", error)
        }
    }

    const findPlayer = () => {
        if (game !== null) {
            for (let i = 0; i < game.players.length; i++) {
                if (game.players[i].id ===  localStorage.getItem("PlayerID")) {
                    return {
                        name: game.players[i].name,
                        character: game.players[i].character,
                        characterType: game.players[i].characterType,
                    }
                }
            }
        }
        return {}
    }

    // 已落座玩家加载
    const sit = () => {
        // 写死，只允许14名玩家
        return Array.from({length: 14}, (_, index) => {
            if (game && game.players[index]) {
                return <span key={index} className="place place-sit">{game.players[index].name}</span>
            }
        })
    }

    return (
        <div id="GAMING" className="GAMING">
            <div className="layout west">
                <div className="layout north">
                    <Button className="btn small-btn" onClick={returnRoom}>返回房间</Button>
                </div>
                <div className="layout my-info">
                    <span>我的名字：<span>{game === null ? "" : findPlayer().name}</span></span>
                    <span>我的身份：<span className="keyword">{game === null ? "" : findPlayer().character}</span></span>
                    <span>身份类型：<span className="keyword">{game === null ? "" : findPlayer().characterType}</span></span>
                </div>
                <span className="layout title">↓ 场上玩家 ↓</span>
                <div className="layout south">
                    {sit()}
                </div>
            </div>
            <div className="layout east" >
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123123123123123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
                <span>123</span>
            </div>
        </div>
    )
}

export default Gaming