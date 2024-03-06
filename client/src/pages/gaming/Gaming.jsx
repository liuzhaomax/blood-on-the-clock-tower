import {useNavigate, useParams} from "react-router-dom"
import {Button, Modal, Switch, notification, Drawer} from "antd"
import {FireOutlined, RollbackOutlined, SmileFilled, SmileOutlined, StopOutlined, ReadOutlined} from "@ant-design/icons"
import React, {useEffect, useState, useMemo} from "react"
import "./Gaming.css"
import {remove} from "../../utils/array"
import {sleep} from "../../utils/time"
import dayImg from "../../assets/images/bg/day.png"
import nightImg from "../../assets/images/bg/night.png"
import "../../utils/captcha/captcha.css"
import {startCaptcha} from "../../utils/captcha/captcha.js"
import Instruction from "./Instruction"
import config from "../../config/config"

const Context = React.createContext({
    name: "Default",
})
let socket
// let timer
let wolfAudio = new Audio("/audio/wolf.wav")
let cockAudio = new Audio("/audio/cock.wav")
let castLock = false // 入夜后给施放技能的时间，后端没有，只在前端限制，因为只限制主机

function Gaming() {
    const navigate = useNavigate()
    let { roomId } = useParams()

    // 加载游戏
    const [game, setGame] = useState(null)
    useEffect(() => {
        establishConn()
    }, [])
    const establishConn = () => {
        // 获取game 长连接
        socket = new WebSocket(`${config.beBaseUrl}/game/${roomId}/${localStorage.getItem("PlayerID")}`)
        socket.onopen = function() {
            loadGame()
        }
        socket.onmessage = function(event) {
            console.log("Received message from server:", JSON.parse(event.data))
            setGame(JSON.parse(event.data))
        }
        socket.onerror = function(error) {
            console.error("WebSocket error:", error)
            establishConn() // 断线重连
        }
    }
    const loadGame = () => {
        let req = JSON.stringify({action: "load_game", targets: []})
        socket.send(req)
    }
    const jumpToHome = () => {
        navigate("/home", {
            replace: true,
            state: "/home",
        })
    }

    // // 防断线
    // useEffect(() => {
    //     timer = setInterval(() => {
    //         keepAlive()
    //     }, 30000)
    // }, [])
    // const keepAlive = () => {
    //     let req = JSON.stringify({action: "ping", targets: []})
    //     socket.send(req)
    // }

    // 退出房间
    const returnRoom = () => {
        showReturnRoomModal()
    }
    const [isReturnRoomModalOpen, setIsReturnRoomModalOpen] = useState(false)
    const showReturnRoomModal = () => {
        setIsReturnRoomModalOpen(true)
    }
    const handleReturnRoomOk = () => {
        castLock = false
        setIsReturnRoomModalOpen(false)
        // clearInterval(timer)
        jumpToHome()
        let req = JSON.stringify({action: "quit_game", targets: []})
        socket.send(req)
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

    // 增加一条log 并上色  addLog(event.data, [/[0-9]/g, "highlight"], ["天", "highlight"])
    let wordClassPairs = [
        [/(?<=第).*?(?=天)|平安夜/g, "highlight highlight-number"], // 数字
        [/\[([^\]]+)]/g, "highlight highlight-player"], // 玩家名字
        [/\{[^}]+}/g, "highlight highlight-skill-result"], // 技能结果关键字
        [/(投毒|卜算|认主|守护|杀害|枪毙|反弹|反向通灵|投给)/g, "highlight highlight-skill"], // 技能关键字
        [/(死亡|处决结果)/g, "highlight highlight-severe"], // 重大事件关键字
        [/(提名)/g, "highlight highlight-nominate"], // 提名
        [/(投票)/g, "highlight highlight-vote"], // 投票
    ]
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
        updateSeatTag()
        updateSeatDead()
    }, [game])
    const loadPersonalLog = async () => {
        if (game) {
            for (let i = 0; i < game.players.length; i++) {
                if (game.players[i].id === localStorage.getItem("PlayerID")) {
                    replaceLog(game.players[i].log, ...wordClassPairs)
                    break
                }
            }
        }
    }

    const getMe = (game) => {
        let me
        for (let i = 0; i < game.players.length; i++) {
            if (game.players[i].id === localStorage.getItem("PlayerID")) {
                me = game.players[i]
                break
            }
        }
        return me
    }

    // 点击玩家名字，选中玩家，保存被选中的玩家ID
    const [selectedPlayers, setSelectedPlayers] = useState([])
    const selectPlayer = (event) => {
        event.preventDefault()
        // 守鸦人可以验死人，占卜可以验死人
        let me = getMe(game)
        if (!event.target.classList.contains("highlight-dead")
            || (me.character === "守鸦人" && !me.state.dead)
            || (me.character === "占卜师" && !me.state.dead)
        ) {
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
    }
    const resetSelectedPlayers = () => {
        let classes = document.getElementsByClassName("seat-selected")
        Array.from(classes).forEach(seat => {
            seat.classList.remove("seat-selected")
        })
        setSelectedPlayers([])
    }

    // 更新seat死亡状态
    const updateSeatDead = () => {
        if (game) {
            let me = getMe(game)
            for (let i = 0; i < game.players.length; i++) {
                if (game.players[i].state.dead) {
                    let seat = document.getElementById(game.players[i].id)
                    if (!seat.classList.contains("highlight-dead")) {
                        seat.classList.add("highlight-dead")
                        seat.classList.add("highlight-dead-cant-be-selected")
                    }
                    if ((me.character === "守鸦人" || me.character === "占卜师")
                        && !me.state.dead
                        && seat.classList.contains("highlight-dead-cant-be-selected")) {
                        seat.classList.remove("highlight-dead-cant-be-selected")
                    }
                    if (!(me.character === "守鸦人" || me.character === "占卜师")
                        && !me.state.dead
                        && seat.classList.contains("seat-selected")) {
                        seat.classList.remove("seat-selected")
                    }
                }
            }
        }
    }

    // 更新seat后面的标签
    const updateSeatTag = () => {
        if (game) {
            let butlerTou
            let tagTou, tagTi, tagBei
            for (let i = 0; i < game.players.length; i++) {
                // 投票标签
                tagTou = document.getElementById(game.players[i].id + "-tou")
                if (game.players[i].ready.vote > 0 && tagTou.classList.contains("tag-hidden")) {
                    tagTou.classList.remove("tag-hidden")
                } else if (game.players[i].ready.vote === 0 && !tagTou.classList.contains("tag-hidden")) {
                    tagTou.classList.add("tag-hidden")
                    if (game.players[i].state.master) {
                        butlerTou = true
                    }
                    if (game.players[i].character === "管家") {
                        tagTou.classList.remove("tag-hidden")
                    }
                }
                // 提名标签
                tagTi = document.getElementById(game.players[i].id + "-ti")
                if (game.players[i].ready.nominate && tagTi.classList.contains("tag-hidden")) {
                    tagTi.classList.remove("tag-hidden")
                } else if (!game.players[i].ready.nominate && !tagTi.classList.contains("tag-hidden")) {
                    tagTi.classList.add("tag-hidden")
                }
                // 被提名标签
                tagBei = document.getElementById(game.players[i].id + "-bei")
                if (game.players[i].ready.nominated && tagBei.classList.contains("tag-hidden")) {
                    tagBei.classList.remove("tag-hidden")
                } else if (!game.players[i].ready.nominated && !tagBei.classList.contains("tag-hidden")) {
                    tagBei.classList.add("tag-hidden")
                }
            }
            if (butlerTou) {
                for (let i = 0; i < game.players.length; i++) {
                    if (game.players[i].character === "管家") {
                        tagTou.classList.add("tag-hidden")
                        break
                    }
                }
            }
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
                                        <span className="individual individual-name" onClick={selectPlayer} id={game.players[index].id}>{game.players[index].name}</span>
                                        <span className="individual-tag tag-tou" id={game.players[index].id + "-tou"}>投</span>
                                        <span className="individual-tag tag-ti" id={game.players[index].id + "-ti"}>提</span>
                                        <span className="individual-tag tag-bei" id={game.players[index].id + "-bei"}>被</span>
                                    </div>
                                )
                            }
                        })
                    }
                </div>
                <Button className="btn mini-btn skill-btn" onClick={cast}><FireOutlined /></Button>
                <div>
                    {
                        Array.from({ length: game ? game.players.length / 2 : 15 / 2 }, (_, index) => {
                            const reversedIndex = game ? game.players.length - 1 - index : 15 - 1 - index
                            if (game && game.players[reversedIndex]) {
                                return (
                                    <div key={index} className="place place-sit seat" id={game.players[reversedIndex].id + "-p"}>
                                        <span className="individual individual-name" onClick={selectPlayer} id={game.players[reversedIndex].id}>{game.players[reversedIndex].name}</span>
                                        <span className="individual-tag tag-tou" id={game.players[reversedIndex].id + "-tou"}>投</span>
                                        <span className="individual-tag tag-ti" id={game.players[reversedIndex].id + "-ti"}>提</span>
                                        <span className="individual-tag tag-bei" id={game.players[reversedIndex].id + "-bei"}>被</span>
                                    </div>
                                )
                            }
                        })
                    }
                </div>
            </div>
        )
    }

    // 跳转复盘页面
    const jumpToReview = () => {
        navigate(`/review/${roomId}`, {
            replace: true,
            state: `/review/${roomId}`,
        })
    }

    // 日夜切换
    useEffect(() => {
        toggleSunMoon()
    }, [game])
    const [iconSunMoon, setIconSunMoon] = useState(true)
    const toggleSunMoon = () => {
        if (game && game.state.night === true) {
            // 重置选择的玩家
            resetSelectedPlayers()
            if (iconSunMoon) {
                setIconSunMoon(false)
                document.getElementById("GAMING").style.backgroundColor = "#35557EFF"
            } else {
                setIconSunMoon(true)
                document.getElementById("GAMING").style.backgroundColor = "#357e5b"
            }
        }
        // 修改背景图片
        updateBgImg()
    }
    const [api, contextHolder] = notification.useNotification()
    const openNotification = (placement) => {
        api.info({
            message: "防连击保护",
            description: <Context.Consumer>{({ name }) => `不好意思, ${name}!`}</Context.Consumer>,
            placement,
        })
    }
    const contextValue = useMemo(
        () => ({
            name: "现在还不能点击，请等待所有技能施放完毕，或等待投票结束",
        }),
        [],
    )

    // 日夜切换按钮
    const toggleNight = () => {
        if (checkReadyToToggleNight()) {
            // 锁定与结算过程
            gameProcess(game.state.stage+1)
        } else {
            openNotification("topRight")
        }
    }
    const checkReadyToToggleNight = () => {
        // 死亡或者已放过技能都是ready
        let ready = true
        for (let i = 0; i < game.players.length; i++) {
            ready = ready && (game.players[i].state.casted || game.players[i].state.dead)
        }
        // 所有有技能的操作完，没技能的点完验证码，时间等待结束，不在投票阶段，则切换日夜，切换后首先结算前一阶段
        return !game.state.votingStep
            && !castLock
            // && ready // TODO 测试用，记得解锁     内侧后删除
            || game.state.stage === 0
    }

    // 游戏过程
    const gameProcess = async (stage) => {
        // TODO 语音- 请大家操作或输入验证码
        if (stage % 2 === 1) {
            wolfAudio.play() // 狼叫
        } else {
            cockAudio.play() // 鸡叫
        }
        if (stage === 1) {
            // 发送日夜切换指令到后端，后端重置状态
            await emitToggleNight()
            // 防抖
            castLock = true
            await sleep(2000)
            castLock = false
        }
        if (stage !== 1 && stage % 2 === 0) {
            // 夜转日，结算前一夜，此时前端stage是双数，但是后端stage依然是单数，因为emitToggleNight还未运行
            await emitCheckoutNight()
            // 发送日夜切换指令到后端，后端重置状态
            await emitToggleNight()
            // 防抖
            castLock = true
            await sleep(2000)
            castLock = false
        }
        if (stage !== 1 && stage % 2 === 1) {
            // 日转夜，结算前一天，此时前端stage是单数，但是后端stage依然是双数，因为emitToggleNight还未运行
            await emitCheckoutDay()
            // 发送日夜切换指令到后端，后端重置状态
            await emitToggleNight()
            // 防抖
            castLock = true
            await sleep(2000)
            castLock = false
        }
    }
    const emitToggleNight = () => {
        let req = JSON.stringify({action: "toggle_night", targets: []})
        socket.send(req) // 会在后端更新stage、night
    }
    const emitCheckoutNight = () => {
        let req = JSON.stringify({action: "checkout_night", targets: []})
        socket.send(req) // 结算本阶段
    }
    const emitCheckoutDay = () => {
        let req = JSON.stringify({action: "checkout_day", targets: []})
        socket.send(req) // 结算本阶段
    }
    const emitCast = () => {
        let req = JSON.stringify({action: "cast", targets: castToPlayersId})
        socket.send(req) // 会在后端更新stage、night
    }
    const emitNominate = () => {
        let req = JSON.stringify({action: "nominate", targets: nominateToPlayersId})
        socket.send(req) // 会在后端更新stage、night
    }
    const emitVote = () => {
        let req = JSON.stringify({action: "vote", targets: []})
        socket.send(req) // 会在后端更新stage、night
    }
    const emitEndVoting = () => {
        let req = JSON.stringify({action: "end_voting", targets: []})
        socket.send(req) // 会在后端更新stage、night
    }

    // 跳转review页面
    useEffect(() => {
        loadAnimation()
    }, [game])
    const loadAnimation = async () => {
        if (game) {
            if (game.result) {
                if (game.result.substring(0, 2) === "平民") {
                    await showGif("Civil-gif", 5000)
                }
                if (game.result.substring(0, 2) === "邪恶") {
                    await showGif("Evil-gif", 5000)
                }
                jumpToReview()
            }
        }
    }
    // 动画 胜利结算
    const showGif = async (id, ms) => {
        let gif = document.getElementById(id)
        gif.classList.add("gif-visible")
        await sleep(ms)
        gif.classList.remove("gif-visible")
    }

    // 提名对象
    const [nominateToPlayersId, setNominateToPlayersId] = useState([])
    // 技能释放对象
    const [castToPlayersId, setCastToPlayersId] = useState([])

    // 提名玩家
    const nominate = () => {
        showNominateModal()
    }
    const [isNominateModalOpen, setIsNominateModalOpen] = useState(false)
    const [nominateModalContent, setNominateModalContent] = useState("抱歉，您此刻无法提名")
    const showNominateModal = () => {
        setIsNominateModalOpen(true)
        setNominateModalContent("抱歉，您此刻无法提名") // 将modal的内容重新初始化，防止错乱
        let me = getMe(game)
        setNominateModalContent(genNominateModalContent(me))
    }
    const handleNominateOk = () => {
        setIsNominateModalOpen(false)
        // 提名的条件是，提名是true，stage是偶数；提名成功的条件是，最快
        let me = getMe(game)
        if (me.ready.nominate && !game.state.night && !game.state.votingStep && game.state.stage !== 0) {
            for (let j = 0; j < game.players.length; j++) {
                if (selectedPlayers[0] === game.players[j].id && game.players[j].ready.nominated) {
                    emitNominate()
                    break
                }
            }
        }
    }
    const handleNominateCancel = () => {
        setIsNominateModalOpen(false)
        setNominateToPlayersId([])
    }

    // 投票玩家
    const vote = () => {
        showVoteModal()
    }
    const [isVoteModalOpen, setIsVoteModalOpen] = useState(false)
    const [voteModalContent, setVoteModalContent] = useState("抱歉，您此刻无法投票")
    const showVoteModal = () => {
        setIsVoteModalOpen(true)
        setVoteModalContent("抱歉，您此刻无法投票") // 将modal的内容重新初始化，防止错乱
        let me = getMe(game)
        setVoteModalContent(genVoteModalContent(me))
    }
    const handleVoteOk = () => {
        setIsVoteModalOpen(false)
        // 投票的条件是，投票是true，stage是偶数，有人被提名game.state.votingStep
        let me = getMe(game)
        if (me.ready.vote && game.state.votingStep && game.state.stage !== 0) {
            for (let j = 0; j < game.players.length; j++) {
                if (game.players[j].state.nominated) {
                    emitVote()
                    break
                }
            }
        }
    }
    const handleVoteCancel = () => {
        setIsVoteModalOpen(false)
    }

    // 发动技能
    const cast = () => {
        showCastModal()
    }
    const [isCastModalOpen, setIsCastModalOpen] = useState(false)
    const [castModalContent, setCastModalContent] = useState("抱歉，您无法发动技能")
    const showCastModal = () => {
        setIsCastModalOpen(true)
        setCastModalContent("抱歉，您无法发动技能") // 将modal的内容重新初始化，防止错乱
        let me = getMe(game)
        if (game.state.stage === 1 &&
            (me.character === "下毒者" ||
            me.character === "占卜师" ||
            me.character === "管家")) {
            setCastModalContent(genCastModalContent(me))
            return
        }
        if (game.state.stage % 2 === 1 && game.state.stage !== 1 &&
            (me.character === "下毒者" ||
                me.character === "僧侣" ||
                me.character === "小恶魔" ||
                me.character === "守鸦人" ||
                me.character === "占卜师" ||
                me.character === "管家")) {
            setCastModalContent(genCastModalContent(me))
            return
        }
        if (game.state.stage % 2 === 0 && me.character === "杀手") {
            setCastModalContent(genCastModalContent(me))
        }
    }
    const handleCastOk = () => {
        setIsCastModalOpen(false)
        // 后端判断 发动技能的条件是，取决于身份，drunk，白天黑夜，还有没有技能；前端随便发动，后端判断成不成功
        let me = getMe(game)
        if (!me.state.casted && !game.state.votingStep && game.state.stage !== 0 &&
            (castToPlayersId.length === 1 &&
                (me.character === "下毒者" ||
                me.character === "管家" ||
                me.character === "僧侣" ||
                me.character === "小恶魔" ||
                me.character === "守鸦人" ||
                me.character === "杀手")
                ||
                (castToPlayersId.length === 2 &&
                me.character === "占卜师")
            )) {
            emitCast()
        }
    }
    const handleCastCancel = () => {
        setIsCastModalOpen(false)
        setCastToPlayersId([])
    }

    // 产生提名Modal的内容
    const genNominateModalContent = (me) => {
        if (game.state.stage === 0) {
            return "本局未开始，不能提名"
        }
        if (me.state.dead) {
            return "您已死亡"
        }
        if (game.state.night) {
            return "夜晚不能提名"
        }
        if (game.state.votingStep) {
            return "投票处决环节不能提名"
        }
        if (!me.ready.nominate) {
            return "您本局已发起过提名"
        }
        if (selectedPlayers.length === 0) {
            return "您想提名不能不选人"
        }
        if (selectedPlayers.length > 1) {
            return "您只能选1人提名"
        }
        if (me.ready.nominate && !game.state.night && !game.state.votingStep) {
            let content = "你确定要在今天的处决中，提名玩家 "
            for (let j = 0; j < game.players.length; j++) {
                if (selectedPlayers[0] === game.players[j].id) {
                    if (!game.players[j].ready.nominated) {
                        return "您选择的玩家 " + "<" + game.players[j].name + "> " + "今日已被提名"
                    }
                    content += "<" + game.players[j].name + "> "
                    setNominateToPlayersId([game.players[j].id])
                    break
                }
            }
            content += "吗？"
            return content
        }
        return "您此刻无法提名"
    }

    // 产生投票Modal的内容
    const genVoteModalContent = (me) => {
        if (game.state.stage === 0) {
            return "本局未开始，不能投票"
        }
        if (!me.ready.vote) {
            return "您本阶段已投过票"
        }
        if (!game.state.votingStep) {
            return "不在投票处决环节不能投票"
        }
        if (me.ready.vote > 0 && game.state.votingStep) {
            let content = "你确定要投票给玩家 "
            for (let j = 0; j < game.players.length; j++) {
                if (game.players[j].state.nominated) {
                    content += "<" + game.players[j].name + "> "
                    break
                }
            }
            content += "吗？"
            return content
        }
        return "您此刻无法投票"
    }

    // 产生技能施放Modal的内容
    const genCastModalContent = (me) => {
        if (game.state.stage === 0) {
            return "本局未开始，不能发动技能"
        }
        if (me.state.dead) {
            return "您已死亡"
        }
        if (game.state.votingStep) {
            return "投票阶段不能发动技能"
        }
        let selectedPlayersObj = []
        let content = "是否要对玩家 "
        for (let i = 0; i < selectedPlayers.length; i++) {
            for (let j = 0; j < game.players.length; j++) {
                if (selectedPlayers[i] === game.players[j].id) {
                    content += "<" + game.players[j].name + "> "
                    selectedPlayersObj.push(game.players[j])
                    break
                }
            }
        }
        switch (me.character) {
        case "下毒者":
            if (!game.state.night) {
                return "白天不能投毒"
            }
            for (let i = 0; i < selectedPlayersObj.length; i++) {
                if (selectedPlayersObj[i].state.dead) {
                    return "您不能毒已死的人"
                }
                if (selectedPlayersObj[i].character === "下毒者") {
                    return "您不能毒自己"
                }
            }
            if (selectedPlayers.length === 1) {
                content += "下毒吗？"
                break
            }
            return "您必须且只能选1个人下毒"
        case "占卜师":
            if (selectedPlayers.length === 2) {
                content += "占卜，看看有没有恶魔吗？"
                break
            }
            return "您只能选2个人占卜"
        case "管家":
            if (!game.state.night) {
                return "白天不能认主"
            }
            for (let i = 0; i < selectedPlayersObj.length; i++) {
                if (selectedPlayersObj[i].character === "管家") {
                    return "您不能认主自己"
                }
            }
            if (selectedPlayers.length === 1) {
                content += "认主？他投你可选投，他不投你投无效。"
                break
            }
            return "您只能选1个人认主"
        case "僧侣":
            if (!game.state.night) {
                return "白天不能守护"
            }
            for (let i = 0; i < selectedPlayersObj.length; i++) {
                if (selectedPlayersObj[i].state.dead) {
                    return "您不能守护已死的人"
                }
                if (selectedPlayersObj[i].character === "僧侣") {
                    return "您不能守护自己"
                }
            }
            if (selectedPlayers.length === 1) {
                content += "进行守护吗？"
                break
            }
            return "您只能选1个人守护"
        case "小恶魔":
            if (!game.state.night) {
                return "白天不能击杀"
            }
            for (let i = 0; i < selectedPlayersObj.length; i++) {
                if (selectedPlayersObj[i].state.dead) {
                    return "您不能杀害已死的人"
                }
            }
            if (selectedPlayers.length === 1) {
                content += "进行杀害吗？"
                break
            }
            return "您只能选1个人杀害"
        case "守鸦人":
            if (!game.state.night) {
                return "白天不能反向通灵"
            }
            if (selectedPlayers.length === 1) {
                content += "进行反向通灵吗？"
                break
            }
            return "您只能选1个人反向通灵"
        case "杀手":
            if (game.state.night) {
                return "夜晚不能开枪"
            }
            if (!me.state.bullet) {
                return "您已没有子弹，不能开枪"
            }
            for (let i = 0; i < selectedPlayersObj.length; i++) {
                if (selectedPlayersObj[i].state.dead) {
                    return "您不能枪毙已死的人"
                }
                if (selectedPlayersObj[i].character === "杀手") {
                    return "您不能枪毙自己"
                }
            }
            if (selectedPlayers.length === 1) {
                content += "实行枪决吗？"
                break
            }
            return "您只能选1个人枪决"
        }
        setCastToPlayersId(selectedPlayers)
        return content
    }

    const endVotingStep = () => {
        if (game && game.state.votingStep) {
            emitEndVoting()
        } else {
            openEndVotingNotification("topRight")
        }
    }
    const openEndVotingNotification = (placement) => {
        api.info({
            message: "非法点击",
            description: <Context.Consumer>{() => "不好意思, 当前不在投票环节，无法结束投票环节!"}</Context.Consumer>,
            placement,
        })
    }

    const [currentStep, setCurrentStep] = useState("本局未开始")
    useEffect(() => {
        loadCurrentStage()
    }, [game])
    const loadCurrentStage = () => {
        if (game) {
            if (game.state.stage === 0) {
                setCurrentStep("本局未开始")
                return
            }
            if (game.state.stage % 2 === 1) {
                setCurrentStep("技能施放")
                return
            }
            if (game.state.stage % 2 === 0 && !game.state.votingStep) {
                setCurrentStep("自由发言")
                return
            }
            if (game.state.stage % 2 === 0 && game.state.votingStep) {
                setCurrentStep("投票处决")
            }
        }
    }

    // 更换背景图片 在toggleSunMoon调用
    const [bgImg, setBgImg] = useState(dayImg)
    const updateBgImg = () => {
        if (game) {
            setBgImg(game.state.stage % 2 === 1 ? nightImg : dayImg)
        }
    }

    // 验证码
    const [isCaptchaModalOpen, setIsCaptchaModalOpen] = useState(false)
    useEffect(() => {
        let container = document.getElementById("CaptchaModal")
        if (container) {
            let box = document.querySelector(".box")
            let titbox = document.querySelector(".tit-box")
            let fls = document.querySelectorAll(".fl")
            let radios = document.querySelectorAll(".radio")
            startCaptcha(box, titbox, fls, radios, container, (success) => {
                setIsCaptchaModalOpen(!success)
            })
        }
    }, [isCaptchaModalOpen])
    useEffect(() => {
        // 弹出验证码
        showCaptchaModal()
    }, [game && game.state.night])
    const showCaptchaModal = () => {
        if (game) {
            let me = getMe(game)
            if (game.state.night && !me.state.dead) {
                let firstNight = false
                if (game.state.stage ===1 ) {
                    firstNight = true
                }
                for (let i = 0; i < game.players.length; i++) {
                    let open = false
                    if (game.players[i].id === localStorage.getItem("PlayerID")) {
                        if (game.players[i].character !== "下毒者"
                            && game.players[i].character !== "占卜师"
                            && game.players[i].character !== "管家") {
                            open = true
                        }
                        if ((game.players[i].character === "小恶魔"
                                || game.players[i].character === "僧侣"
                                || game.players[i].character === "守鸦人")
                            && firstNight) {
                            open = true
                        }
                        if ((game.players[i].character === "小恶魔"
                                || game.players[i].character === "僧侣"
                                || game.players[i].character === "守鸦人")
                            && !firstNight) {
                            open = false
                        }
                        if (open) {
                            setIsCaptchaModalOpen(true)
                            break
                        }
                    }
                }
            } else {
                setIsCaptchaModalOpen(false)
            }
        }
    }
    // TODO 测试代码 开始   内侧后删除
    const handleCaptchaCancel = () => {
        setIsCaptchaModalOpen(false)
    }
    // TODO 测试代码 结束   内侧后删除

    // 游戏说明
    const [open, setOpen] = useState(false)
    const showDrawer = () => {
        setOpen(true)
    }
    const onClose = () => {
        setOpen(false)
    }

    return (
        <div id="GAMING" className="GAMING" style={{backgroundImage: `url(${bgImg})`}}>
            <img className="gif" id="Evil-gif" src={process.env.PUBLIC_URL + "/video/evil.gif"} alt="Evil GIF"/>
            <img className="gif" id="Civil-gif" src={process.env.PUBLIC_URL + "/video/civil.gif"} alt="Civil GIF"/>
            <div className="layout west">
                <div className="layout north">
                    <Button className="btn small-btn" onClick={returnRoom}><RollbackOutlined /></Button>
                    <Switch className="switch" checkedChildren="显示身份" unCheckedChildren="隐藏身份" defaultChecked onChange={checkSwitch} />
                    <Button className="btn small-btn" onClick={showDrawer}><ReadOutlined /></Button>
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
                        <Button className="btn mini-btn" onClick={nominate}>提名</Button>
                        <Button className="btn mini-btn" onClick={vote}>投票</Button>
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
            <Context.Provider value={contextValue}>
                {contextHolder}
            </Context.Provider>
            <Modal title="验证码" open={isCaptchaModalOpen} onCancel={handleCaptchaCancel} footer={null}>
                <div id="CaptchaModal">
                    {isCaptchaModalOpen ?
                        <div className="verify-box">
                            <div className="box">
                                <div className="bg-blur"></div>
                            </div>
                            <div className="tit-box"></div>
                        </div>
                        :
                        <></>
                    }
                </div>
            </Modal>
            <Drawer
                title="游戏说明书"
                placement="top"
                closable={false}
                onClose={onClose}
                open={open}
                key="top"
            >
                <Instruction/>
            </Drawer>
        </div>
    )
}

export default Gaming
