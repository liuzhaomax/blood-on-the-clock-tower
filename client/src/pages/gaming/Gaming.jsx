import {useNavigate, useParams} from "react-router-dom"
import {Button, Switch} from "antd"
import { RollbackOutlined, SmileFilled, FireOutlined } from "@ant-design/icons"
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
        const socket = new WebSocket(`ws://192.168.1.14:8080/game/${roomId}`)
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

    const checkSwitch = () => {
        let myInfo = document.getElementsByClassName("keyword")
        for (let i = 0; i < myInfo.length; i++) {
            if (myInfo[i].style.visibility === "hidden") {
                myInfo[i].style.visibility = "visible"
            } else {
                myInfo[i].style.visibility = "hidden"
            }
        }
    }

    return (
        <div id="GAMING" className="GAMING">
            <div className="layout west">
                <div className="layout north">
                    <Button className="btn small-btn" onClick={returnRoom}><RollbackOutlined /></Button>
                    <Switch className="switch" checkedChildren="显示身份" unCheckedChildren="隐藏身份" defaultChecked onChange={checkSwitch} />
                    <Button className="btn small-btn" onClick={returnRoom}><SmileFilled /></Button>
                </div>
                <div className="layout info-wrap">
                    <div className="layout my-info">
                        <span>我的名字：<span>{game === null ? "" : findPlayer().name}</span></span>
                        <span>我的身份：<span className="keyword">{game === null ? "" : findPlayer().character}</span></span>
                        <span>身份类型：<span className="keyword">{game === null ? "" : findPlayer().characterType}</span></span>
                        <Button className="mini-btn" onClick={returnRoom}>提名</Button>
                        <Button className="mini-btn" onClick={returnRoom}>投票</Button>
                        <Button className="mini-btn skill-btn" onClick={returnRoom}><FireOutlined /></Button>
                    </div>
                    <span className="layout title">↓ 场上玩家 ↓</span>
                    <div className="layout south">
                        {sit()}
                    </div>
                </div>
            </div>
            <div className="layout east" >
                <span>未开始，尚未入第一夜</span>
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