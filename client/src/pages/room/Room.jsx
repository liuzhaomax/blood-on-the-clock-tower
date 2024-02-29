import React, {useEffect, useMemo, useState} from "react"
import "./Room.css"
import {Button, Flex, notification} from "antd"
import {useNavigate, useParams} from "react-router-dom"
import config from "../../config/config"

const Context = React.createContext({
    name: "Default",
})

let socket

function Room() {
    const navigate = useNavigate()
    let { roomId } = useParams()
    const [room, setRoom] = useState(null)

    useEffect(() => {
        establishConn()
    }, [])
    const establishConn = () => {
        socket = new WebSocket(`${config.beBaseUrl}/room/${roomId}`)
        socket.onopen = function() {
            loadRoom()
        }
        socket.onmessage = function(event) {
            // console.log("Received message from server:", JSON.parse(event.data))
            setRoom(JSON.parse(event.data))
            checkToStartGame(JSON.parse(event.data))
        }
        socket.onerror = function(error) {
            console.error("WebSocket error:", error)
        }
    }
    const loadRoom = () => {
        let data = {
            action: "list_players",
            payload: localStorage.getItem("PlayerID"),
        }
        socket.send(JSON.stringify(data))
    }

    const checkToStartGame = (room) => {
        switch (room.status) {
        case "游戏中":
            jumpToGame()
            break
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
        let data = {
            action: "quit_room",
            payload: localStorage.getItem("PlayerID"),
        }
        socket.send(JSON.stringify(data))
        jumpToHome()
    }
    const jumpToHome = () => {
        navigate("/home", {
            replace: true,
            state: "/home",
        })
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

    const jumpToGame = () => {
        navigate(`/game/${roomId}`, {
            replace: true,
            state: `/game/${roomId}`,
        })
    }

    const startGame = () => {
        // 小于5人，无法开始游戏
        if (room.players.length < 5) {
            openPlayerNumNotification("topRight")
            return
        }
        // 多于15人，无法开始游戏
        if (room.players.length > 15) {
            openPlayerNum2Notification("topRight")
            return
        }
        // 有人在复盘页面，无法开始游戏
        if (room && room.status === "复盘中") {
            openReviewingNotification("topRight")
            return
        }
        let data = {
            action: "start_game",
            payload: "",
        }
        socket.send(JSON.stringify(data))
    }
    const [api, contextHolder] = notification.useNotification()
    const openReviewingNotification = (placement) => {
        api.info({
            message: "房间未准备好",
            description: <Context.Consumer>{() => "不好意思, 有玩家尚未退出复盘，请督促返回房间!"}</Context.Consumer>,
            placement,
        })
    }
    const openPlayerNumNotification = (placement) => {
        api.info({
            message: "房间未准备好",
            description: <Context.Consumer>{() => "人数不足五人，无法开始游戏!"}</Context.Consumer>,
            placement,
        })
    }
    const openPlayerNum2Notification = (placement) => {
        api.info({
            message: "房间未准备好",
            description: <Context.Consumer>{() => "人数多于十五人，无法开始游戏!"}</Context.Consumer>,
            placement,
        })
    }
    const contextValue = useMemo(
        () => ({
            name: "",
        }),
        [],
    )

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
                <div className="layout seat-wrap-wrap">
                    {sit()}
                </div>
            </div>
            <Context.Provider value={contextValue}>
                {contextHolder}
            </Context.Provider>
        </div>
    )
}

export default Room
