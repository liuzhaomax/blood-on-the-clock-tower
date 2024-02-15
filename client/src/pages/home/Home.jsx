import React from "react"
import "./Home.css"
import { HOME } from "../../config/cstModule"
import { Button, Flex, List, Avatar } from "antd"

const data = [
    {
        title: "Ant Design Title 1",
    },
    {
        title: "Ant Design Title 2",
    },
    {
        title: "Ant Design Title 3",
    },
    {
        title: "Ant Design Title 4",
    },
]

function Home() {
    return (
        <div id={HOME.KEY} className={HOME.KEY}>
            <Flex className="layout" wrap="wrap">
                <Button className="btn">创建房间</Button>
            </Flex>
            <div className="layout">
                <List style={{ width: "100%" }}
                    itemLayout="horizontal"
                    dataSource={data}
                    renderItem={(item, index) => (
                        <List.Item className="list-item">
                            <List.Item.Meta
                                avatar={<Avatar src={`https://api.dicebear.com/7.x/miniavs/svg?seed=${index}`} />}
                                title={<a href="https://ant.design">{item.title}</a>}
                                description="Ant Design, a design language for background applications, is refined by Ant UED Team"
                            />
                        </List.Item>
                    )}
                />
            </div>
        </div>
    )
}

export default Home