// 建立WebSocket连接
const socket = new WebSocket("ws://localhost:8080/websocket")

// 当连接打开时执行的回调函数
socket.onopen = function() {
    console.log("WebSocket connected!")
    // 发送消息给后端
    socket.send("Hello, Server!")
}

// 当接收到消息时执行的回调函数
socket.onmessage = function(event) {
    console.log("Received message from server:", event.data)
}

// 当发生错误时执行的回调函数
socket.onerror = function(error) {
    console.error("WebSocket error:", error)
}

// 当连接关闭时执行的回调函数
socket.onclose = function(event) {
    console.log("WebSocket closed:", event)
}

export default socket
