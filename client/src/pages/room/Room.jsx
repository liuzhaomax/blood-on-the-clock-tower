import React from "react"
import "./Room.css"
import { Flex, Button } from "antd"

function Room() {

    return (
        <div id="ROOM" className="ROOM">
            <Flex className="layout" horizontal="true" gap="middle" justify="space-evenly" align="center" wrap="wrap">
                <Button className="btn" type="dashed">abc</Button>
            </Flex>
            <Flex className="layout" horizontal="true" gap="middle" justify="space-evenly" align="center" wrap="wrap">
                de
            </Flex>
        </div>
    )
}

export default Room
