import React, {useMemo} from "react"
import { useState, useEffect } from "react"
import "./Home.css"
import { HOME } from "../../config/cstModule"
import { Button, Flex, List, Avatar, Drawer, Space, Input, notification } from "antd"
import {genShortUUID} from "../../utils/uuid"
import { useNavigate } from "react-router-dom"
import {blood} from "../../utils/blood/blood"
import {sleep} from "../../utils/time"
import config from "../../config/config"
import ICP from "../../assets/images/ICP.png"
import bat from "../../assets/video/bat.gif"

if (localStorage.getItem("PlayerID") === null) {
    localStorage.setItem("PlayerID", genShortUUID())
}
if (localStorage.getItem("PlayerName") === null) {
    localStorage.setItem("PlayerName", "好人1号-" + genShortUUID().slice(-6))
}

const Context = React.createContext({
    name: "Default",
})

let socketHome

function Home() {
    const [roomList, setRoomList] = useState([])

    useEffect(() => {
        establishConn()
    }, [])
    const establishConn = () => {
        socketHome = new WebSocket(`${config.beBaseUrl}/home`)
        socketHome.onopen = function() {
            loadRoomList()
        }
        socketHome.onmessage = function(event) {
            // console.log("Received message from server:", JSON.parse(event.data))
            setRoomList(JSON.parse(event.data))
        }
        socketHome.onerror = function(error) {
            console.error("WebSocket error:", error)
        }
    }
    const loadRoomList = () => {
        let data = {
            action: "list_rooms",
            payload: localStorage.getItem("PlayerID"),
        }
        socketHome.send(JSON.stringify(data))
    }
    const navigate = useNavigate()
    const jump = roomId => {
        navigate(`/room/${roomId}`, {
            replace: true,
            state: { roomId }
        })
    }

    const [open, setOpen] = useState(false)
    const showDrawer = () => {
        setOpen(true)
        setPlayerName(localStorage.getItem("PlayerName"))
    }
    const onClose = () => {
        setOpen(false)
    }

    const [open1, setOpen1] = useState(false)
    const showDrawer1 = room => {
        setOpen1(true)
        setRoomId(room.id)
        setRoomName(room.name)
        setRoomSelected(room)
        setPlayerName(localStorage.getItem("PlayerName"))
    }
    const onClose1 = () => {
        setOpen1(false)
    }

    const [roomSelected, setRoomSelected] = useState(null)

    const createRoom = () => {
        // 房间名称不能为空
        if (roomName === "") {
            openRoomNameNotification("topRight")
            return
        }
        onClose()
        let player = {}
        player.id = localStorage.getItem("PlayerID")
        player.name = playerName
        player.waiting = true
        let roomInfo = {
            id: roomId,
            name: roomName,
            password: roomPassword,
            host: localStorage.getItem("PlayerID"),
            players: [player],
        }
        let data = {
            action: "create_room",
            payload: roomInfo,
        }
        socketHome.send(JSON.stringify(data))
        jump(roomInfo.id)
    }
    
    const [roomId, setRoomId] = useState(genShortUUID())
    const [roomName, setRoomName] = useState("房间-" + genShortUUID().slice(-6))
    const [roomPassword, setRoomPassword] = useState("")
    const [playerName, setPlayerName] = useState("好人1号")

    const handleRoomNameChange = (event) => {
        setRoomName(event.target.value)
    }

    const handleRoomPasswordChange = (event) => {
        setRoomPassword(event.target.value)
    }

    const handlePlayerNameChange = (event) => {
        setPlayerName(event.target.value)
        localStorage.setItem("PlayerName", event.target.value)
    }

    const joinRoom = () => {
        // 如果房间status是游戏中，则无法加入
        if (roomSelected && (roomSelected.status === "游戏中" || roomSelected.status === "复盘中")) {
            openGamingNotification("topRight")
            return
        }
        // 人数大于等于15，则无法加入房间
        if (roomSelected && roomSelected.players.length >= 15) {
            openPlayerNum2Notification("topRight")
            return
        }
        onClose1()
        let playerInfo = {
            id: localStorage.getItem("PlayerID"),
            name: playerName,
            waiting: true,
        }
        let roomInfo = {
            id: roomId,
            password: roomPassword,
        }
        let data = {
            action: "join_room",
            payload: {
                room: roomInfo,
                player: playerInfo,
            },
        }
        socketHome.send(JSON.stringify(data))
        jump(roomId)
    }
    const [api, contextHolder] = notification.useNotification()
    const openPlayerNum2Notification = (placement) => {
        api.info({
            message: "操作无效",
            description: <Context.Consumer>{() => "人数多于十五人，无法加入房间!"}</Context.Consumer>,
            placement,
        })
    }
    const openGamingNotification = (placement) => {
        api.info({
            message: "操作无效",
            description: <Context.Consumer>{() => "游戏已开始或还在复盘，无法加入房间!"}</Context.Consumer>,
            placement,
        })
    }
    const openRoomNameNotification = (placement) => {
        api.info({
            message: "信息确实",
            description: <Context.Consumer>{() => "房间名称不能为空!"}</Context.Consumer>,
            placement,
        })
    }
    const contextValue = useMemo(
        () => ({
            name: "",
        }),
        [],
    )

    // 动画 蝙蝠 流血
    useEffect(() => {
        blood()
        hideGif("Bat-gif", 3000)
    }, [])
    const hideGif = async (id, ms) => {
        let gif = document.getElementById(id)
        await sleep(ms)
        if (gif) {
            gif.classList.add("hidden")
        }
    }
    
    return (
        <div id={HOME.KEY} className={HOME.KEY}>
            <div id="Title-wrap">
                <div id="Title">血染钟楼</div>
            </div>
            <svg className="svg">
                <filter id="noise">
                    <feTurbulence baseFrequency="0.07" type="fractalNoise" result="turbNoise"></feTurbulence>
                    <feDisplacementMap in="SourceGraphic" in2="turbNoise" xChannelSelector="G" yChannelSelector="B" scale="6" result="disp"></feDisplacementMap>
                </filter>
            </svg>
            {window.innerWidth >= 958 ?
                <img id="Bat-gif" src={bat} alt="Bat GIF"/>
                :
                <></>
            }
            <Flex className="layout" wrap="wrap">
                <Button className="btn-main" onClick={showDrawer}>创建房间</Button>
            </Flex>
            <div className="layout">
                { roomList ?
                    <List style={{ width: "100%" }}
                        itemLayout="horizontal"
                        dataSource={roomList}
                        renderItem={(item, index) => (
                            <List.Item className="list-item">
                                <List.Item.Meta
                                    avatar={<Avatar src={`https://api.dicebear.com/7.x/miniavs/svg?seed=${index}`} />}
                                    title={
                                        <span>{item.name}</span>
                                    }
                                    description={
                                        <Flex horizontal="true" gap="middle" justify="space-between" align="center" wrap="wrap">
                                            <span>当前人数：{item.players === null ? 0 : item.players.length}</span>
                                            <span>{item.status}</span>
                                        </Flex>
                                    }
                                    onClick={showDrawer1.bind(this, item)}
                                />
                            </List.Item>
                        )}
                    />
                    :
                    <div></div>
                }
            </div>
            <div id="Warn-health">
                <p>健康游戏忠告</p>
                <p>抵制不良游戏，拒绝盗版游戏。</p>
                <p>注意自我保护，谨防受骗上当。</p>
                <p>适度游戏益脑，沉迷游戏伤身。</p>
                <p>合理安排时间，享受健康生活。</p>
            </div>
            <div id="ICP">
                <img src={ICP} alt="ICP"/>
                <a href="https://beian.mps.gov.cn/#/query/webSearch" target="\_blank" style={{color: "white",textDecoration:"none"}}>
                    津ICP备2024011617号
                </a>
            </div>
            <Drawer
                title="创建房间"
                placement={"bottom"}
                width={500}
                onClose={onClose}
                open={open}
                extra={
                    <Space>
                        <Button className="small-btn" onClick={onClose}>取消</Button>
                        <Button className="small-btn" type="primary" onClick={createRoom}>确定</Button>
                    </Space>
                }
            >
                <Space direction="vertical" size="middle">
                    <p>房间名称</p>
                    <Space.Compact>
                        <Input className="input" value={roomName} onChange={handleRoomNameChange} />
                    </Space.Compact>
                    <p>房间密码</p>
                    <Space.Compact>
                        <Input className="input" value={roomPassword} onChange={handleRoomPasswordChange} />
                    </Space.Compact>
                    <p>我的名字</p>
                    <Space.Compact>
                        <Input className="input" value={playerName} onChange={handlePlayerNameChange} />
                    </Space.Compact>
                </Space>
            </Drawer>
            <Drawer
                title="加入房间"
                placement={"bottom"}
                width={500}
                onClose={onClose1}
                open={open1}
                extra={
                    <Space>
                        <Button className="small-btn" onClick={onClose1}>取消</Button>
                        <Button className="small-btn" type="primary" onClick={joinRoom}>确定</Button>
                    </Space>
                }
            >
                <Space direction="vertical" size="middle">
                    <p>房间名称</p>
                    <Space.Compact>
                        <Input className="input" value={roomName} onChange={handleRoomNameChange} />
                    </Space.Compact>
                    <p>房间密码</p>
                    <Space.Compact>
                        <Input className="input" value={roomPassword} onChange={handleRoomPasswordChange} />
                    </Space.Compact>
                    <p>我的名字</p>
                    <Space.Compact>
                        <Input className="input" value={playerName} onChange={handlePlayerNameChange} />
                    </Space.Compact>
                </Space>
            </Drawer>
            <Context.Provider value={contextValue}>
                {contextHolder}
            </Context.Provider>
        </div>
    )
}

export default Home