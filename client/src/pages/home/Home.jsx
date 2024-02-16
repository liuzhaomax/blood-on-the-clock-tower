import React from "react"
import { useState, useEffect } from "react"
import "./Home.css"
import { HOME } from "../../config/cstModule"
import { Button, Flex, List, Avatar, Drawer, Space, Input } from "antd"
import {genShortUUID} from "../../utils/uuid"
import { useNavigate } from "react-router-dom"

if (localStorage.getItem("PlayerID") === "") {
    localStorage.setItem("PlayerID", genShortUUID())
}

function Home() {
    const navigate = useNavigate()
    const jumpToPath = roomId => {
        navigate(`/room/${roomId}`, {
            replace: true,
            state: { roomId }
        })
    }

    const [open, setOpen] = useState(false)
    const showDrawer = () => {
        setOpen(true)
    }
    const onClose = () => {
        setOpen(false)
    }

    const [open1, setOpen1] = useState(false)
    const showDrawer1 = room => {
        setOpen1(true)
        setRoomId(room.id)
        setRoomName(room.name)
    }
    const onClose1 = () => {
        setOpen1(false)
    }

    const [roomList, setRoomList] = useState(null)

    useEffect(() => {
        setTimeout(() => {
            loadRoomList()
        }, 500)
    })

    const loadRoomList = () => {
        const socket = new WebSocket("ws://localhost:8080/home")
        socket.onopen = function() {
            socket.send("list rooms")
        }
        socket.onmessage = function(event) {
            // console.log("Received message from server:", JSON.parse(event.data))
            if (event.data !== null) {
                setRoomList(JSON.parse(event.data))
            }
        }
        socket.onerror = function(error) {
            console.error("WebSocket error:", error)
        }
    }

    const createRoom = () => {
        onClose()
        let player = {}
        player.id = localStorage.getItem("PlayerID")
        player.name = playerName
        let roomInfo = {
            id: roomId,
            name: roomName,
            password: roomPassword,
            state: "等待开始",
            players: [player],
        }
        const socket = new WebSocket("ws://localhost:8080/createRoom")
        socket.onopen = function() {
            console.log("WebSocket for creating room connected!")
            socket.send(JSON.stringify(roomInfo))
        }
        jumpToPath(roomInfo.id)
    }
    
    const [roomId, setRoomId] = useState(genShortUUID())
    const [roomName, setRoomName] = useState("房间-" + genShortUUID())
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
    }

    const joinRoom = () => {
        onClose1()
        let playerInfo = {
            id: localStorage.getItem("PlayerID"),
            name: playerName,
        }
        const socket = new WebSocket(`ws://localhost:8080/joinRoom/${roomId}/${roomPassword}`)
        socket.onopen = function() {
            socket.send(JSON.stringify(playerInfo))
        }
        socket.onmessage = function(event) {
            // console.log("Received message from server:", JSON.parse(event.data))
            if (event.data !== null) {
                jumpToPath(roomId)
            }
            // TODO else 弹出密码错误对话框
        }
        socket.onerror = function(error) {
            console.error("WebSocket error:", error)
        }
    }
    
    return (
        <div id={HOME.KEY} className={HOME.KEY}>
            <Flex className="layout" wrap="wrap">
                <Button className="btn" onClick={showDrawer}>创建房间</Button>
            </Flex>
            <div className="layout">
                { roomList !== null ?
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
                                            <span>{item.state}</span>
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
        </div>
    )
}

export default Home