import "./Review.css"
import {useNavigate, useParams} from "react-router-dom"
import {Button} from "antd"
import React, {useEffect, useState} from "react"
import config from "../../config/config"

function Review() {
    const navigate = useNavigate()
    let { roomId } = useParams()

    const [game, setGame] = useState(null)
    useEffect(() => {
        setTimeout(() => {
            loadGame()
            checkToReturnRoom()
        }, 100)
    }, [game])
    const loadGame = () => {
        const socket = new WebSocket(`${config.beBaseUrl}/review/${roomId}`)
        socket.onopen = function() {
            socket.send("review")
        }
        socket.onmessage = function(event) {
            // console.log("Received message from server:", JSON.parse(event.data))
            setGame(JSON.parse(event.data))
        }
        socket.onerror = function(error) {
            console.error("WebSocket error:", error)
        }
    }
    const checkToReturnRoom = () => {
        if (game) {
            switch (game.status) {
            case "等待开始":
                jumpToRoom()
                break
            }
        }
    }

    // 返回房间
    const returnRoom = () => {
        let meId = localStorage.getItem("PlayerID")
        const socket = new WebSocket(`${config.beBaseUrl}/returnRoom/${roomId}/${meId}`)
        socket.onopen = function() {
            socket.send("return_room")
            jumpToRoom()
        }
        socket.onerror = function(error) {
            console.error("WebSocket error:", error)
        }
    }
    const jumpToRoom = () => {
        navigate(`/room/${roomId}`, {
            replace: true,
            state: `/room/${roomId}`,
        })
    }

    // 加载总日志
    useEffect(() => {
        loadPersonalLog()
    }, [game])
    const loadPersonalLog = () => {
        if (game) {
            replaceLog(game.log, ...wordClassPairs)
        }
    }
    let wordClassPairs = [
        [/(?<=第).*?(?=天)|平安夜|\(.*?\)/g, "highlight highlight-number"], // 数字
        [/\[([^\]]+)]/g, "highlight highlight-player"], // 玩家名字
        [/\{[^}]+}/g, "highlight highlight-skill-result"], // 技能结果关键字
        [/(投毒|卜算|认主|守护|杀害|枪毙|弹|反向通灵|投给)/g, "highlight highlight-skill"], // 技能关键字
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

    return (
        <div id="REVIEW" className="REVIEW">
            <div className="layout north">
                <Button className="btn small-btn" onClick={returnRoom}>返回房间</Button>
            </div>
            <div className="layout central">
                {game && game.result === "邪恶阵营胜利" ?
                    <div className="win-evil">邪恶阵营胜利</div>
                    :
                    <div className="win-civil">平民阵营胜利</div>
                }
            </div>
            <div className="layout title">
                复盘信息
            </div>
            <div className="layout south" id="LOG">
                {game ? game.log : "总日志为空"}
            </div>
        </div>
    )
}

export default Review