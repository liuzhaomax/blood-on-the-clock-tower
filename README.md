# Blood on the ClockTower 血染钟楼

血染钟楼桌游程序化流水线，相当于说书人角色

## 需求与设计

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
    + 加载结算 /checkout/:roomId
        + 输入：无
        + 输出：无
    + 返回房间 /return/:roomId
        + 输入：无
        + 输出：无
+ 其他功能
    + 房间15分钟没有开始，自动销毁
    + 等待开始时，断线重连 √
    + 游戏中时，断线重连


## 日志设计

### 日志分类
+ 总日志：记录全部
+ 个人日志：记录自己可见的事件

### 事件分类
+ 公共事件（透明底，白字）
    + 重大事件
        + 样式：红字
        + 词语：死亡，被处决，提名
    + 玩家名字
        + 样式：黄底，黑字
        + 词语：好人1号
    + 数字
        + 样式：紫色
        + 词语：第1夜
+ 私有事件
    + 个人技能
        + 样式：蓝字
        + 词语：下毒，查验，杀害
    + 玩家名字
        + 样式：黄底，黑字
        + 词语：好人1号
    + 数字
        + 样式：紫色
        + 词语：第1夜


## 游戏流程

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
+ 游戏中时，退出游戏会怎样
+ 退出游戏后，主机转变

