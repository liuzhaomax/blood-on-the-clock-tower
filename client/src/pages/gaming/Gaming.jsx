import {useNavigate, useParams} from "react-router-dom"
import {Button, Modal, Switch} from "antd"
import {FireOutlined, RollbackOutlined, SmileFilled, SmileOutlined} from "@ant-design/icons"
import React, {useEffect, useState} from "react"
import "./Gaming.css"

let Night = false
let socketGaming

// 防断线
setInterval(() => {
    if (socketGaming) {
        socketGaming.send("ping")
    }
}, 10000)

function Gaming() {
    const navigate = useNavigate()
    let { roomId } = useParams()
    const [game, setGame] = useState(null)

    // 加载游戏
    useEffect(() => {
        loadGame()
    }, [])
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

    // 退出房间
    const returnRoom = () => {
        showModal()
    }
    const [isModalOpen, setIsModalOpen] = useState(false)
    const showModal = () => {
        setIsModalOpen(true)
    }
    const handleOk = () => {
        setIsModalOpen(false)
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
    const handleCancel = () => {
        setIsModalOpen(false)
    }

    // 显示玩家信息
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

    // 显示隐藏身份
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

    // 日夜切换
    useEffect(() => {
        if (game && game.night !== Night) {
            toggleSunMoon()
            Night = game.night
        }
    }, [game])
    const [iconSunMoon, setIconSunMoon] = useState(true)
    const toggleSunMoon = () => {
        if (iconSunMoon) {
            setIconSunMoon(false)
            document.getElementById("GAMING").style.backgroundColor = "#35557EFF"
        } else {
            setIconSunMoon(true)
            document.getElementById("GAMING").style.backgroundColor = "#357e5b"
        }
    }
    const toggleNight = () => {
        socketGaming.send("toggle_night")
        loadGame()
    }

    // 建立长连接
    useEffect(() => {
        gamingConn()
    }, [])
    const gamingConn = () => {
        socketGaming = new WebSocket(`ws://192.168.1.14:8080/gaming/${roomId}/${localStorage.getItem("PlayerID")}`)
        socketGaming.onmessage = function(event) {
            // console.log("Received message from server:", event.data)
            addLog(
                event.data,
                [/[0-9]/g, "highlight-number"], // 数字
            )
        }
        socketGaming.onerror = function(error) {
            console.error("WebSocket error:", error)
            gamingConn() // 断线重连
        }
    }

    // 增加一条log 并上色  addLog(event.data, [/[0-9]/g, "highlight"], ["天", "highlight"])
    const addLog = (text, ...wordClassPairs) => {
        let replacedText = replaceText(text, ...wordClassPairs[0])
        if (wordClassPairs.length > 1) {
            for (let i = 1; i < wordClassPairs.length; i++) {
                replacedText = replaceText(replacedText, ...wordClassPairs[i])
            }
        }
        if (document.getElementById("LOG")) {
            document.getElementById("LOG").innerHTML = document.getElementById("LOG").innerHTML + `<span>${replacedText}</span>`
        }
    }
    const replaceText = (text, word, className) => {
        if (typeof word === "string") {
            let regex = new RegExp(word, "g")
            return text.replace(regex, `<span class=${className}>${word}</span>`)
        }
        return text.replace(word, match => `<span class=${className}>${match}</span>`)
    }

    // 已落座玩家加载
    const sit = () => {
        return(
            <div className="south">
                <div>
                    {
                        Array.from({ length: game ? game.players.length / 2 : 15 / 2 }, (_, index) => {
                            if (game && game.players[index]) {
                                return (
                                    <div key={index} className="place place-sit seat">
                                        {game.players[index].name}
                                    </div>
                                )
                            }
                        })
                    }
                </div>
                <Button className="mini-btn skill-btn" onClick={returnRoom}><FireOutlined /></Button>
                <div>
                    {
                        Array.from({ length: game ? game.players.length / 2 : 15 / 2 }, (_, index) => {
                            const reversedIndex = game ? game.players.length - 1 - index : 15 - 1 - index
                            if (game && game.players[reversedIndex]) {
                                return (
                                    <div key={reversedIndex} className="place place-sit seat">
                                        {game.players[reversedIndex].name}
                                    </div>
                                )
                            }
                        })
                    }
                </div>
            </div>
        )
    }

    return (
        <div id="GAMING" className="GAMING">
            <div className="layout west">
                <div className="layout north">
                    <Button className="btn small-btn" onClick={returnRoom}><RollbackOutlined /></Button>
                    <Switch className="switch" checkedChildren="显示身份" unCheckedChildren="隐藏身份" defaultChecked onChange={checkSwitch} />
                    { game && game.host === localStorage.getItem("PlayerID")
                        ?
                        <Button className="btn small-btn" onClick={toggleNight}>{iconSunMoon ? <SmileFilled /> : <SmileOutlined />}</Button>
                        :
                        <></>
                    }
                </div>
                <div className="layout info-wrap">
                    <div className="layout my-info">
                        <span>我的名字：<span>{game === null ? "" : findPlayer().name}</span></span>
                        <span>我的身份：<span className="keyword">{game === null ? "" : findPlayer().character}</span></span>
                        <span>身份类型：<span className="keyword">{game === null ? "" : findPlayer().characterType}</span></span>
                        <Button className="mini-btn" onClick={returnRoom}>提名</Button>
                        <Button className="mini-btn" onClick={returnRoom}>投票</Button>
                    </div>
                    <span className="layout title">↓ 场上玩家 ↓</span>
                    {sit()}
                </div>
            </div>
            <div className="layout east" id="LOG" >
                <span>未开始，尚未入第一夜</span>
            </div>
            <Modal title="退出游戏" open={isModalOpen} onOk={handleOk} onCancel={handleCancel}>
                <p>退出游戏后，不可重新进入游戏中的房间，确定退出请点击“OK”</p>
            </Modal>
        </div>
    )
}

export default Gaming
