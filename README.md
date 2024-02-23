# 血染钟楼 Blood on the ClockTower 

> 血染钟楼桌游线下辅助器

**亮点**：
+ 取代说书人角色，所有线下小伙伴都可以深度参与解谜
+ 全程无需闭眼，不会有长时间的闭目养神和睁眼闭眼过程
+ 全程无倒计时，玩家们可以自由控制各阶段时长
+ 无需按序施放技能，玩家们可同时在某夜或某日释放技能
+ 操作简单，所有操作最多只需点3下，例如施放技能：选人 → 点技能按钮(每人只有一个) → 点确定
+ 丰富明确的提示，解开大部分对于是否可操作的疑惑，减少解答时间
+ 多彩高亮的日志，帮助玩家记录
+ 防场外，即便晚上没技能，也需要操作手机；隐藏身份信息等

**不足**：
+ 目前只支持《血染钟楼：灾祸滋生》版本
+ 不提供说书人角色，就想当主持人的小伙伴不好意思了
+ 防场外不是绝对的，不过也可以利用场外来掩盖真相，比如故意操作很长时间

## 1. 说明书
### 1.1 游戏概述
+ 游戏版本：灾祸滋生
+ 阵营与角色
    + 村民（平民阵营）
        + 洗衣妇
        + 图书管理员
        + 调查员
        + 厨师
        + 共情者
        + 占卜师
        + 掘墓人
        + 僧侣
        + 守鸦人
        + 圣女
        + 杀手
        + 士兵
        + 市长
    + 外来者（平民阵营）
        + 管家
        + 酒鬼
        + 隐士
        + 圣徒
    + 爪牙（邪恶阵营）
        + 下毒者
        + 间谍
        + 魅魔
        + 男爵
    + 恶魔（邪恶阵营）
        + 小恶魔

> 具体技能见正版游戏说明书

+ 胜利条件
    + 平民：恶魔死亡
    + 邪恶
        + 场上没人可以提名且(没有枪手或有枪手没有子弹)
        + 票数不大于活人的半数，且邪恶玩家数量不小于活人的半数
        + 平民阵营被屠城
        + 圣徒被投票处决

### 1.2 游戏流程（结合辅助器）

> 全程无需闭眼

初始状态
[图片]
点击右上角日夜切换按钮进入第一夜
[图片]
选择玩家，发动技能
[图片]
点击右上角日夜切换按钮进入白天
[图片]
点击玩家，发动提名
[图片]
点击右上角结束投票按钮，结算投票结果
[图片]
如果没有人被处决，则还可以提名
[图片]
大家都同意进入夜晚，则点击日夜切换按钮
[图片]
游戏结束后，进入复盘页面
[图片]

### 1.3 规则细节
+ 下毒者可以毒任何人，包括小恶魔


## 2. 需求与设计

页面
----功能 API
--------输入输出

+ 房间列表页 √
    + 加载房间列表 /home 
        + 输入：无
        + 输出：roomList
    + 创建房间 /createRoom 
        + 功能：创建的房间的state赋值
        + 输入：无
        + 输出：无
    + 加入房间 /joinRoom/:roomId/:roomPassword 
        + 功能：验证密码；当前用户重连；新玩家加入
        + 输入：player{id, name}
        + 输出：无
+ 等待开始页 √
    + 加载房间 /room/:roomId
        + 输入：无
        + 输出：room
    + 退出房间 /quitRoom/:roomId
        + 功能：房间减少当前玩家，如果房间人数为0，则销毁房间；
        + 输入：player{id}
        + 输出：无
    + 开始游戏 /startGame/:roomId
        + 功能：改变房间state
        + 输入：无
        + 输出：无
+ 游戏中页
    + 加载游戏 /game/:roomId
        + 功能：根据人数，分配身份；夜晚读秒
        + 输入：无
        + 输出：无
    + 游戏过程 /gaming/:roomId/:playerId
        + 输入：无
        + 输出：操作结果日志
+ 游戏结算页
    + 加载结算 /review/:roomId
        + 输入：无
        + 输出：无
    + 返回房间 /returnRoom/:roomId
        + 输入：无
        + 输出：无
+ 其他功能
    + 房间15分钟没有开始，自动销毁
    + 等待开始时，断线重连 √
    + 游戏中时，断线重连


## 3. 日志设计

### 日志分类
+ 总日志：记录全部
+ 个人日志：记录自己可见的事件

### 事件分类
+ 公共事件（透明底，白字）
    + 重大事件
        + 样式：红字加粗
        + 词语：死亡|处决结果|提名|投票|平安夜
    + 玩家名字
        + 样式：黄底，黑字
        + 词语：[]包裹
    + 数字
        + 样式：紫色
        + 词语：第1夜
+ 私有事件
    + 个人技能
        + 样式：蓝字
        + 词语：下毒|占卜|认主|守护|杀害|枪毙|弹
    + 技能结果
        + 样式：绿底黑字
        + 词语：{}包裹


## 4. 游戏流程设计

+ 阶段 stage：夜晚和白天各为一个阶段
+ 环节 step：行为互斥
    + 技能施放环节：夜晚发动技能的角色在此环节施放技能
    + 自由发言环节：白天发动技能的角色在此环节施放技能，此环节可提名
    + 投票处决环节：此环节可投票

### 第一夜
1. 下毒 操作
2. 洗衣妇 日志
3. 图书管理员 日志
4. 调查员 日志
5. 厨师 日志
6. 共情者 日志
7. 占卜师 操作
8. 管家 操作
9. 间谍 日志

### 其他夜晚
1. 下毒 操作
2. 僧侣 操作
3. 魅魔 日志
4. 小恶魔 操作
5. 守鸦人 日志
6. 共情者 日志
7. 占卜师 操作
8. 管家 操作
9. 掘墓人 日志

### 白天
1. 圣女 日志
2. 杀手 操作
3. 士兵 日志
4. 市长 日志
5. 圣徒 日志




## TODO
+ 直接通过链接进入房间，不被允许
+ 玩家名字不允许重复
+ 标记玩家，自认他是什么，我认为他是什么
+ 游戏中时，退出游戏不允许再进入
+ 退出游戏后，主机转变
+ 人少不能开始游戏

