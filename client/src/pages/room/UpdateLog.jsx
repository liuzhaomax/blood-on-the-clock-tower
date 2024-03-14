import React from "react"
import {marked} from "marked"

const text = `
v1.0.4 更新内容：
1. 在房间页面增加游戏说明与更新日志按钮，可以相应查看

v1.0.3 更新内容：
1. 间谍重做，符合wiki
2. 修改了bug，现在少于7人，不会提供小恶魔三个不在场的村民
3. 由于间谍的改动，调查员有可能发现没有爪牙
4. 由于间谍的改动，图书管理员可能看到注册为酒鬼的间谍，以注册身份为标准
5. 由于间谍的改动，下毒者毒间谍不仅仅切换间谍为邪恶状态，同时将间谍注册为爪牙-间谍，魔典亦有被毒惩罚
6. 由于间谍的改动，如果间谍被注册为村民，依然可以被圣女反弹死
7. 由于间谍的改动，守鸦人与掘墓人看到的是间谍被注册的身份，除非他们被毒
8. 由于间谍的改动，爪牙恶魔互认身份时，将不能看到间谍的身份，只能看到是爪牙

v1.0.2 更新内容：
1. 修复了进入“复盘中”房间的bug
2. 全部玩家返回房间后，会更新首页房间状态

v1.0.1 更新内容：
1. 夜间技能必须施放，不然无法切换日夜
2. 防场外功能测试程序取消，验证码必须验证通过
`

function UpdateLog() {
    const htmlContent = marked(text)

    return (
        <div id="UPDATELOG" className="UPDATELOG">
            <div dangerouslySetInnerHTML={{ __html: htmlContent }}></div>
        </div>
    )
}

export default UpdateLog
