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
+ 等待开始页
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
        + 输入：无
        + 输出：无
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


## TODO
+ 直接通过链接进入房间，不被允许
+ 玩家名字不允许重复
+ 标记玩家，自认他是什么，我认为他是什么

