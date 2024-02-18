import {useNavigate, useParams} from "react-router-dom"
import {Button, Modal, Switch} from "antd"
import {FireOutlined, RollbackOutlined, SmileFilled, SmileOutlined} from "@ant-design/icons"
import React, {useEffect, useState} from "react"
import "./Gaming.css"
import {remove} from "../../utils/array"

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
        showReturnRoomModal()
    }
    const [isReturnRoomModalOpen, setIsReturnRoomModalOpen] = useState(false)
    const showReturnRoomModal = () => {
        setIsReturnRoomModalOpen(true)
    }
    const handleReturnRoomOk = () => {
        setIsReturnRoomModalOpen(false)
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
    const handleReturnRoomCancel = () => {
        setIsReturnRoomModalOpen(false)
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
        setTimeout(() => {
            toggleSunMoon()
        }, 50)
    }, [game])
    const [iconSunMoon, setIconSunMoon] = useState(true)
    const toggleSunMoon = () => {
        if (game && game.night !== Night) {
            if (iconSunMoon) {
                setIconSunMoon(false)
                document.getElementById("GAMING").style.backgroundColor = "#35557EFF"
            } else {
                setIconSunMoon(true)
                document.getElementById("GAMING").style.backgroundColor = "#357e5b"
            }
            Night = game.night
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
                [/[0-9]/g, "highlight highlight-number"], // 数字
            )
            loadGame()
        }
        socketGaming.onerror = function(error) {
            console.error("WebSocket error:", error)
            gamingConn() // 断线重连
        }
    }

    // 增加一条log 并上色  addLog(event.data, [/[0-9]/g, "highlight"], ["天", "highlight"])
    const addLog = (text, ...wordClassPairs) => {
        let replacedText = updateText(text, ...wordClassPairs[0])
        if (wordClassPairs.length > 1) {
            for (let i = 1; i < wordClassPairs.length; i++) {
                replacedText = updateText(replacedText, ...wordClassPairs[i])
            }
        }
        if (document.getElementById("LOG")) {
            document.getElementById("LOG").innerHTML = document.getElementById("LOG").innerHTML + `<span>${replacedText}</span>`
        }
    }
    const updateText = (text, word, className) => {
        if (typeof word === "string") {
            let regex = new RegExp(word, "g")
            return text.replace(regex, `<span class="${className}">${word}</span>`)
        }
        return text.replace(word, match => `<span class="${className}">${match}</span>`)
    }
    const replaceLog = (text, ...wordClassPairs) => {
        let replacedText = updateText(text, ...wordClassPairs[0])
        if (wordClassPairs.length > 1) {
            for (let i = 1; i < wordClassPairs.length; i++) {
                replacedText = updateText(replacedText, ...wordClassPairs[i])
            }
        }
        let removedNTextArr = replacedText.split("\n").map(item => {
            return `<span>${item}</span>`
        })
        let resultText = removedNTextArr.join("")
        if (document.getElementById("LOG")) {
            document.getElementById("LOG").innerHTML = `${resultText}`
        }
    }

    // 刷新页面日志加载
    useEffect(() => {
        loadPersonalLog()
    }, [game])
    const loadPersonalLog = () => {
        if (game) {
            for (let i = 0; i < game.players.length; i++) {
                if (game.players[i].id === localStorage.getItem("PlayerID")) {
                    replaceLog(
                        game.players[i].log,
                        [/[0-9]/g, "highlight highlight-number"], // 数字
                    )
                    break
                }
            }
        }
    }

    // 点击玩家名字，选中玩家，保存被选中的玩家ID
    let selectedPlayers = []
    const selectPlayer = (event) => {
        event.preventDefault()
        if (event.target.classList.contains("seat-selected")) {
            event.target.classList.remove("seat-selected")
            remove(selectedPlayers, event.target.id)
        } else {
            event.target.classList.add("seat-selected")
            selectedPlayers.push(event.target.id)
        }
    }

    // 已落座玩家加载
    const sit = () => {
        return(
            <div className="south">
                <div>
                    {
                        Array.from({ length: game ? Math.ceil(game.players.length / 2) : Math.ceil(15 / 2) }, (_, index) => {
                            if (game && game.players[index]) {
                                return (
                                    <div key={index} className="place place-sit seat">
                                        <span className="individual" onClick={selectPlayer} id={game.players[index].id}>{game.players[index].name}</span>
                                        <span className="individual-tag tag-tou">投</span>
                                        <span className="individual-tag tag-ti">提</span>
                                        <span className="individual-tag tag-bei">被</span>
                                    </div>
                                )
                            }
                        })
                    }
                </div>
                <Button className="mini-btn skill-btn" onClick={cast}><FireOutlined /></Button>
                <div>
                    {
                        Array.from({ length: game ? game.players.length / 2 : 15 / 2 }, (_, index) => {
                            const reversedIndex = game ? game.players.length - 1 - index : 15 - 1 - index
                            if (game && game.players[reversedIndex]) {
                                return (
                                    <div key={index} className="place place-sit seat">
                                        <span className="individual" onClick={selectPlayer} id={game.players[index].id}>{game.players[reversedIndex].name}</span>
                                        <span className="individual-tag tag-tou">投</span>
                                        <span className="individual-tag tag-ti">提</span>
                                        <span className="individual-tag tag-bei">被</span>
                                    </div>
                                )
                            }
                        })
                    }
                </div>
            </div>
        )
    }

    // 提名玩家
    const nominate = () => {
        showNominateModal()
    }
    const [isNominateModalOpen, setIsNominateModalOpen] = useState(false)
    const showNominateModal = () => {
        setIsNominateModalOpen(true)
    }
    const handleNominateOk = () => {
        setIsNominateModalOpen(false)
    }
    const handleNominateCancel = () => {
        setIsNominateModalOpen(false)
    }

    // 投票玩家
    const vote = () => {
        showVoteModal()
    }
    const [isVoteModalOpen, setIsVoteModalOpen] = useState(false)
    const showVoteModal = () => {
        setIsVoteModalOpen(true)
    }
    const handleVoteOk = () => {
        setIsVoteModalOpen(false)
    }
    const handleVoteCancel = () => {
        setIsVoteModalOpen(false)
    }

    // 发动技能
    const cast = () => {
        showCastModal()
    }
    const [isCastModalOpen, setIsCastModalOpen] = useState(false)
    const showCastModal = () => {
        setIsCastModalOpen(true)
    }
    const handleCastOk = () => {
        setIsCastModalOpen(false)
    }
    const handleCastCancel = () => {
        setIsCastModalOpen(false)
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
                        <Button className="mini-btn" onClick={nominate}>提名</Button>
                        <Button className="mini-btn" onClick={vote}>投票</Button>
                    </div>
                    <span className="layout title">↓ 场上玩家 ↓</span>
                    {sit()}
                </div>
            </div>
            <div className="layout east" id="LOG" >
                <span>未开始，尚未入第一夜</span>
            </div>
            <Modal title="退出游戏" open={isReturnRoomModalOpen} onOk={handleReturnRoomOk} onCancel={handleReturnRoomCancel}>
                <p>退出游戏后，不可重新进入游戏中的房间，确定退出请点击“OK”</p>
            </Modal>
            <Modal title="提名" open={isNominateModalOpen} onOk={handleNominateOk} onCancel={handleNominateCancel}>
                <p>提名</p>
            </Modal>
            <Modal title="投票" open={isVoteModalOpen} onOk={handleVoteOk} onCancel={handleVoteCancel}>
                <p>投票</p>
            </Modal>
            <Modal title="发动技能" open={isCastModalOpen} onOk={handleCastOk} onCancel={handleCastCancel}>
                <p>投票</p>
            </Modal>
        </div>
    )
}

export default Gaming
