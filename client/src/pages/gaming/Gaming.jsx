import {useNavigate, useParams} from "react-router-dom"
import {Button, Modal, Switch} from "antd"
import {FireOutlined, RollbackOutlined, SmileFilled, SmileOutlined, StopOutlined} from "@ant-design/icons"
import React, {useEffect, useState} from "react"
import "./Gaming.css"
import {remove} from "../../utils/array"
import {sleep} from "../../utils/time"

let socketGaming
let wolfAudio = new Audio("/audio/wolf.wav")
let GameState = {
    night: false, // 是否是黑夜
    stage: 0, // 偶数是白天，奇数是黑夜，stage==1是特殊的第一夜
    castingStep: false, // 是否处于释放技能阶段，即讨论阶段，白天晚上同理
    votingStep: false, // 是否有人被提名，也就是是否进入投票环节
    checkoutStep: true, // 是否处于对当前阶段的结算环节，对于初始化，只有在这个阶段才能点击切换日夜
}

// 防断线
setInterval(() => {
    if (socketGaming) {
        socketGaming.send("ping")
    }
}, 10000)

function Gaming() {
    const navigate = useNavigate()
    let { roomId } = useParams()

    // 加载游戏
    const [game, setGame] = useState(null)
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
        GameState = {
            night: false,
            stage: 0,
            castingStep: false,
            votingStep: false,
            checkoutStep: true,
        }
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
    const [selectedPlayers, setSelectedPlayers] = useState([])
    const selectPlayer = (event) => {
        event.preventDefault()
        let selectedPlayersCopy = selectedPlayers.slice()
        if (event.target.classList.contains("seat-selected")) {
            event.target.classList.remove("seat-selected")
            remove(selectedPlayersCopy, event.target.id)
            setSelectedPlayers(selectedPlayersCopy)
        } else {
            event.target.classList.add("seat-selected")
            selectedPlayersCopy.push(event.target.id)
            setSelectedPlayers(selectedPlayersCopy)
        }
    }
    const resetSelectedPlayers = () => {
        let classes = document.getElementsByClassName("seat-selected")
        Array.from(classes).forEach(seat => {
            seat.classList.remove("seat-selected")
        })
        setSelectedPlayers([])
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

    // 日夜切换
    useEffect(() => {
        setTimeout(() => {
            toggleSunMoon()
            updateGameState()
        }, 50)
    }, [game])
    const [iconSunMoon, setIconSunMoon] = useState(true)
    const toggleSunMoon = () => {
        if (game && game.state.night !== GameState.night) {
            if (iconSunMoon) {
                setIconSunMoon(false)
                document.getElementById("GAMING").style.backgroundColor = "#35557EFF"
            } else {
                setIconSunMoon(true)
                document.getElementById("GAMING").style.backgroundColor = "#357e5b"
            }
            GameState.night = game.state.night
        }
    }
    const updateGameState = () => {
        if (game && game.state.night !== GameState.night) {
            GameState.night = game.state.night
        }
        if (game && game.state.stage !== GameState.stage) {
            GameState.stage = game.state.stage
        }
        if (game && game.state.castingStep !== GameState.castingStep) {
            GameState.castingStep = game.state.castingStep
        }
        if (game && game.state.votingStep !== GameState.votingStep) {
            GameState.votingStep = game.state.votingStep
        }
        if (game && game.state.checkoutStep !== GameState.checkoutStep) {
            GameState.checkoutStep = game.state.checkoutStep
        }
    }
    const toggleNight = () => {
        // 重置选择的玩家
        resetSelectedPlayers()
        if (GameState.checkoutStep) {
            GameState.stage++
            if (GameState.stage % 2 === 1) {
                setCurrentStep("施放技能")
            } else {
                setCurrentStep("自由发言")
            }
            GameState.checkoutStep = false
            GameState.castingStep = true
            socketGaming.send("toggle_night")
            loadGame()
            process()
        } else {
            console.log("不是当前stage的结算环节")
        }
    }

    // 游戏过程
    const process = async () => {
        if (GameState.stage === 1) {
            // 初始化状态
            GameState.checkoutStep = false
            // 狼叫
            wolfAudio.play()
            // 语音- 请大家操作或输入验证码
            // 等1秒
            // 弹出验证码
            // 等9秒
            await sleep(2000)
            // 所有有技能的操作完，没技能的点完验证码，时间等待结束，则进入结算步骤  GameState.checkoutStep = true
            console.log("是1")
        }
        if (GameState.stage !== 1 && GameState.stage % 2 === 0) {
            console.log("是偶数")
        }
        if (GameState.stage !== 1 && GameState.stage % 2 === 1) {
            console.log("是奇数")
        }
    }

    // 提名玩家
    const nominate = () => {
        showNominateModal()
    }
    const [isNominateModalOpen, setIsNominateModalOpen] = useState(false)
    const [nominateModalContent, setNominateModalContent] = useState("抱歉，您此刻无法提名")
    const showNominateModal = () => {
        setIsNominateModalOpen(true)
        let me
        for (let i = 0; i < game.players.length; i++) {
            if (game.players[i].id === localStorage.getItem("PlayerID")) {
                me = game.players[i]
                break
            }
        }
        if (GameState.castingStep && !me.ready.nominated) {
            setNominateModalContent(genNominateModalContent(me))
        }
    }
    const handleNominateOk = () => {
        setIsNominateModalOpen(false)
        // 提名的条件是，提名是true，stage是偶数
        // 提名成功的条件是，最快
    }
    const handleNominateCancel = () => {
        setIsNominateModalOpen(false)
    }

    // 投票玩家
    const vote = () => {
        showVoteModal()
    }
    const [isVoteModalOpen, setIsVoteModalOpen] = useState(false)
    const [voteModalContent, setVoteModalContent] = useState("抱歉，您此刻无法投票")
    const showVoteModal = () => {
        setIsVoteModalOpen(true)
        let me
        for (let i = 0; i < game.players.length; i++) {
            if (game.players[i].id === localStorage.getItem("PlayerID")) {
                me = game.players[i]
                break
            }
        }
        if (GameState.votingStep && !me.ready.voted) {
            setVoteModalContent(genVoteModalContent(me))
        }
    }
    const handleVoteOk = () => {
        setIsVoteModalOpen(false)
        // 投票的条件是，投票是true，stage是偶数，有人被提名GameState.votingStep
    }
    const handleVoteCancel = () => {
        setIsVoteModalOpen(false)
    }

    // 发动技能
    const cast = () => {
        showCastModal()
    }
    const [isCastModalOpen, setIsCastModalOpen] = useState(false)
    const [castModalContent, setCastModalContent] = useState("抱歉，您此刻无法发动技能或没有主动技能")
    const showCastModal = () => {
        setIsCastModalOpen(true)
        let me
        for (let i = 0; i < game.players.length; i++) {
            if (game.players[i].id === localStorage.getItem("PlayerID")) {
                me = game.players[i]
                break
            }
        }
        if (GameState.castingStep && !me.ready.casted) {
            if (GameState.stage === 1 &&
                (me.character === "下毒者" ||
                me.character === "占卜师" ||
                me.character === "管家")) {
                setCastModalContent(genCastModalContent(me))
                return
            }
            if (GameState.stage % 2 === 1 && GameState.stage !== 1 &&
                (me.character === "下毒者" ||
                    me.character === "僧侣" ||
                    me.character === "小恶魔" ||
                    me.character === "占卜师" ||
                    me.character === "管家")) {
                setCastModalContent(genCastModalContent(me))
                return
            }
            if (GameState.stage % 2 === 0 && me.character === "杀手") {
                setCastModalContent(genCastModalContent(me))
            }
        }
    }
    const handleCastOk = () => {
        setIsCastModalOpen(false)
        // 发动技能的条件是，取决于身份，drunk，白天黑夜，还有没有技能
    }
    const handleCastCancel = () => {
        setIsCastModalOpen(false)
    }

    // 产生提名Modal的内容
    const genNominateModalContent = (me) => {
        if (me.nominate) {
            let content = "你确定要在今天的处决中，提名玩家 "
            for (let i = 0; i < selectedPlayers.length; i++) {
                for (let j = 0; j < game.players.length; j++) {
                    if (selectedPlayers[i] === game.players[j].id) {
                        content += "<" + game.players[j].name + "> "
                        break
                    }
                }
            }
            content += "吗？"
            return content
        }
        return "抱歉，您此刻无法提名"
    }

    // 产生投票Modal的内容
    const genVoteModalContent = (me) => {
        if (me.character === "管家") {
            return "您只能跟随主人投票"
        }
        if (me.vote > 0) {
            let content = "你确定要投票给玩家 "
            for (let i = 0; i < selectedPlayers.length; i++) {
                for (let j = 0; j < game.players.length; j++) {
                    if (selectedPlayers[i] === game.players[j].id) {
                        content += "<" + game.players[j].name + "> "
                        break
                    }
                }
            }
            content += "吗？"
            return content
        }
        return "抱歉，您此刻无法投票"
    }

    // 产生技能施放Modal的内容
    const genCastModalContent = (me) => {
        let content = "是否要对玩家 "
        for (let i = 0; i < selectedPlayers.length; i++) {
            for (let j = 0; j < game.players.length; j++) {
                if (selectedPlayers[i] === game.players[j].id) {
                    content += "<" + game.players[j].name + "> "
                    break
                }
            }
        }
        switch (me.character) {
        case "下毒者":
            if (selectedPlayers.length === 1) {
                content += "下毒吗？"
                break
            }
            return "您只能选1个人下毒"
        case "占卜师":
            if (selectedPlayers.length === 2) {
                content += "占卜，看看有没有恶魔吗？"
                break
            }
            return "您只能选2个人占卜"
        case "管家":
            if (selectedPlayers.length === 1) {
                content += "跟随，让他这轮有两票吗？"
                break
            }
            return "您只能选1个人跟随"
        case "僧侣":
            if (selectedPlayers.length === 1) {
                content += "进行守护吗？"
                break
            }
            return "您只能选1个人守护"
        case "小恶魔":
            if (selectedPlayers.length === 1) {
                content += "进行杀害吗？"
                break
            }
            return "您只能选1个人杀害"
        case "杀手":
            if (selectedPlayers.length === 1) {
                content += "实行枪决吗？"
                break
            }
            return "您只能选1个人枪决"
        }
        return content
    }

    const endVotingStep = () => {
        GameState.votingStep = false
        // 这里检查是进哪个环节
        // GameState.castingStep = true
        // GameState.checkoutStep = true
    }

    const [currentStep, setCurrentStep] = useState("未开始")

    return (
        <div id="GAMING" className="GAMING">
            <div className="layout west">
                <div className="layout north">
                    <Button className="btn small-btn" onClick={returnRoom}><RollbackOutlined /></Button>
                    <Switch className="switch" checkedChildren="显示身份" unCheckedChildren="隐藏身份" defaultChecked onChange={checkSwitch} />
                    { game && game.host === localStorage.getItem("PlayerID")
                        ?
                        <>
                            <Button className="btn small-btn" onClick={toggleNight}>{iconSunMoon ? <SmileFilled /> : <SmileOutlined />}</Button>
                            <Button className="btn small-btn" onClick={endVotingStep}><StopOutlined /></Button>
                        </>
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
                    <span className="layout title">当前环节：{currentStep}</span>
                    {sit()}
                </div>
            </div>
            <div className="layout east" id="LOG" >
                <span>未开始，尚未入第一夜</span>
            </div>
            <Modal title="退出游戏" open={isReturnRoomModalOpen} onOk={handleReturnRoomOk} onCancel={handleReturnRoomCancel}>
                <p>退出游戏后，不可重新进入游戏中的房间，确定退出请点击“OK”</p>
            </Modal>
            <Modal id="NominateModal" title="提名" open={isNominateModalOpen} onOk={handleNominateOk} onCancel={handleNominateCancel}>
                <p>{nominateModalContent}</p>
            </Modal>
            <Modal id="VoteModal" title="投票" open={isVoteModalOpen} onOk={handleVoteOk} onCancel={handleVoteCancel}>
                <p>{voteModalContent}</p>
            </Modal>
            <Modal id="CastModal" title="发动技能" open={isCastModalOpen} onOk={handleCastOk} onCancel={handleCastCancel}>
                <p>{castModalContent}</p>
            </Modal>
        </div>
    )
}

export default Gaming
