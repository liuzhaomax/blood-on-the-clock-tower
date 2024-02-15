import React from "react"
import { useState, useEffect } from "react"
import "./Home.css"
import { HOME } from "../../config/cstModule"
import { Button, Flex, List, Avatar, Drawer, Space, Input } from "antd"
import {genShortUUID} from "../../utils/uuid"
import { useNavigate } from "react-router-dom"

const PlayID = genShortUUID()

function Home() {
    const navigate = useNavigate()
    const jumpToPath = path => {
        navigate(path, {
            replace: true,
            state: { path }
        })
    }

    const [open, setOpen] = useState(false)
    const showDrawer = () => {
        setOpen(true)
    }
    const onClose = () => {
        setOpen(false)
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

    const onClick = () => {
        onClose()
        let player = {}
        player.id = PlayID
        let roomInfo = {
            id: genShortUUID(),
            name: roomName,
            password: roomPassword,
            state: "等待开始",
            players: [player],
        }
        const socket = new WebSocket("ws://localhost:8080/room")
        socket.onopen = function() {
            console.log("WebSocket for creating room connected!")
            socket.send(JSON.stringify(roomInfo))
        }
        jumpToPath("/room/" + roomInfo.id)
    }
    
    const [roomName, setRoomName] = useState(genShortUUID())
    const [roomPassword, setRoomPassword] = useState("")

    const handleRoomNameChange = (event) => {
        setRoomName(event.target.value)
    }

    const handleRoomPasswordChange = (event) => {
        setRoomPassword(event.target.value)
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
                                    title={<a href="https://ant.design">{item.name}</a>}
                                    description={
                                        <Flex horizontal="true" gap="middle" justify="space-between" align="center" wrap="wrap">
                                            <span>当前房间人数：{item.players === null ? 0 : item.players.length}</span>
                                            <span>{item.state}</span>
                                        </Flex>
                                    }
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
                        <Button className="small-btn" type="primary" onClick={onClick}>确定</Button>
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
                </Space>
            </Drawer>
        </div>
    )
}

export default Home