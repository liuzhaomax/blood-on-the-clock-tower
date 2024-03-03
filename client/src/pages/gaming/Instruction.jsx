import React from "react"
import {marked} from "marked"

const text = `
## 1. 说明书
### 1.1 游戏概述

> 具体技能见正版游戏说明书，和1.3节（技能细节）

+ 游戏版本：灾祸滋生
+ 阵营与角色
    + 村民（平民阵营）：洗衣妇、图书管理员、调查员、厨师、共情者、占卜师、掘墓人、僧侣、守鸦人、圣女、杀手、士兵、市长
    + 外来者（平民阵营）：管家、酒鬼、隐士、圣徒
    + 爪牙（邪恶阵营）：下毒者、间谍、魅魔、男爵
    + 恶魔（邪恶阵营）：小恶魔
+ 胜利条件
    + 平民：
        1. 恶魔被铲除
        2. 白天剩三人，其中一个是市长，且当日无人被处决
    + 邪恶
        1. 场上没有平民可以提名，且没有杀手或有杀手没有子弹，且没有市长或市长已死
        2. 达成邪恶胜利条件二：可投的票数不大于活人的半数（向上取整），且存活的邪恶玩家数量不小于活人的半数（向上取整），且没有杀手或有杀手没有子弹，且没有市长或市长已死
        3. 平民阵营被屠城
        4. 圣徒被投票处决
+ 与无辅助器的不同之处
    + 没有说书人
    + 全程不需闭眼
    + 技能施放没有顺序限制
    + 玩家面板上，玩家是围坐成一个圈的，也就是左右区域顶部两个玩家是邻座，左右区域底部两个玩家是邻座
    + 间谍（详情见1.3节）
    + 隐士（详情见1.3节）
    + 守鸦人除第一夜，每晚都要点验。假设自己死亡，你要验谁（详情见1.3节）
+ 建议规则
    + 自觉，不要给别人看屏幕，私聊也不要给看，更不要发送截图，这个算作弊
    + 允许公聊，一对一私聊（允许发微信），不要一对多或者多对多私聊
    + 虽然有一定的防窥屏功能，但是还是建议操作时端起手机，平常用纸张衣物遮挡屏幕
    + 等大家都确认已施放技能时，主机再点击切换日夜
    + 等大家都确认不再提名时，主机再点击切换日夜
    + 等大家都确认不再投票时，主机再点击结算投票
+ 注意事项
    + 由于没有登录，一个浏览器绑定一名玩家
    + 五小时未开始游戏，或开始游戏超过五小时未再开始游戏，则房间销毁，如果长时间不开始游戏，请先返回首页再进行操作
    + 如果不在本局中，请不要尝试通过连接进入房间
    + 如果发现操作不正常，且不属于五小时未操作的情况，则刷新页面


### 1.2 游戏流程（结合辅助器）

> 全程无需闭眼，技能施放无顺序

1. 初始状态
[图片]
2. 点击右上角日夜切换按钮进入第一夜
[图片]
3. 选择玩家，发动技能
[图片]
4. 点击右上角日夜切换按钮进入白天
[图片]
5. 点击玩家，发动提名
[图片]
6. 点击右上角结束投票按钮，结算投票结果
[图片]
7. 如果没有人被处决，则还可以提名
[图片]
8. 大家都同意进入夜晚，则点击日夜切换按钮
[图片]
9. 游戏结束后，进入复盘页面
[图片]

### 1.3 技能细节
+ 洗衣妇
    + 被毒后，有极小概率获取真实信息
+ 图书管理员
    + 被毒后，有极小概率获取真实信息
    + 被毒后，也有极小概率获取“没有外来者”
+ 调查员
    + 被毒后，有极小概率获取真实信息
+ 厨师
    + 一对是两恶相连，两对是三恶连座
+ 共情者
    + 邻座不算死人
+ 占卜师
    + 只会提示有和没有，不会提供数量
    + 隐士被视作恶魔，占卜会读到“有”
    + 有一个玩家会在开局前就设定好被占卜视为的恶魔身份不会变
    + 占卜视作的恶魔和隐士被大家当作恶魔是独立的
    + 可以验死人，死人会受其他技能影响，例如死了的被当作恶魔的隐士被毒，会验不出恶魔
+ 掘墓人
    + 挖到隐士，会看到他被当作的身份，而不是隐士；如果隐士中毒，看到的是隐士
    + 隐士被处决，下毒再毒，挖出来是隐士身份，而不是被当作的身份
+ 僧侣
    + 不能守自己
    + 守护可挡刀，不可挡毒
+ 守鸦人
    + 守鸦人需要在每晚（除第一夜）验人（程序需要），可以每晚假定死的是自己，则要验谁
    + 转天自己没死，则没有信息
    + 可以验死人
    + 验到隐士，会看到他被当作的身份，而不是隐士；如果隐士中毒，看到的是隐士
+ 圣女
    + 村民提名立即触发技能
    + 其他身份提名，会让圣女失去能力，而且进入投票环节
    + 被毒，也会消耗技能，但不会有人被弹死
+ 杀手
    + 白天开枪，立即结算
    + 枪毙被当作恶魔的隐士，隐士直接死亡
    + 被枪毙者不暴露身份
+ 士兵
    + 可与守护叠加
+ 市长
    + 可达成平民胜利条件二
+ 管家
    + 主人投票，他可选投
    + 主人不投票，他票作废，会再结算投票时把票补回来
    + 管家不能认主死人
    + 中毒后，依然收到认主信息，但实际没有认主，本轮不可能投票有效
+ 酒鬼
    + 开局获取身份，自己并不知道自己是酒鬼
+ 隐士
    + 开局被当作的邪恶身份，爪牙或恶魔
    + 如果被当作间谍，别人看他，他将不是邪恶的
    + 被毒会变回隐士
+ 圣徒
    + 可达成邪恶胜利条件四
+ 下毒者
    + 毒隐士会让隐士不再邪恶，如果隐士被视作恶魔，被毒也不再被视作恶魔
    + 毒间谍会让间谍变为邪恶
    + 毒只在当晚和接下来的白天起作用
    + 死在夜里，或者转变为恶魔，当夜技能无效
+ 魅魔
    + 恶魔自刀，优先变恶魔
    + 恶魔被处决后，活人数量依然大于等于5，转变为恶魔
+ 男爵
    + 开局技能生效
+ 间谍
    + 第一夜可获取所有身份，身份绝对正确，获取身份不受毒的影响
    + 被毒会变为邪恶
    + 恶魔可以直接知道间谍的身份，而不是只知道他是爪牙
+ 小恶魔
    + 被毒就是平安夜
    + 有一定概率直接刀死市长，市长会有概率转移弹死除恶魔外的一切活人
    + 自刀，有魅魔，魅魔先变小恶魔，没魅魔，随机一个变

### 1.4 进入游戏UI介绍
从上到下，从左到右
+ 主控区域
    + 返回房间按钮：一旦退出游戏，不可再重新进入刚才退出的游戏
    + 显示/隐藏身份开关：防窥屏
    + 游戏功能介绍按钮
    + 日夜切换按钮：带有防连击保护功能
    + 结束投票按钮：只能在投票环节点击
+ 主信息
    + 名字
    + 身份
    + 身份类型
    + 提名按钮
    + 投票按钮
+ 玩家面板
    + 当前环节：未开始、技能施放、自由发言、投票处决
    + 玩家名字
    + 技能按钮
+ 日志面板
`

function Instruction() {
    const htmlContent = marked(text)

    return (
        <div id="INSTRUCTION" className="INSTRUCTION">
            <div dangerouslySetInnerHTML={{ __html: htmlContent }}></div>
        </div>
    )
}

export default Instruction
