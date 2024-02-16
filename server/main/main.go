package main

import (
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/handler"
	"log"
	"net/http"
)

func main() {
	// 房间列表页
	http.HandleFunc("/home", handler.ListRooms)
	http.HandleFunc("/createRoom", handler.CreateRoom)
	http.HandleFunc("/joinRoom/", handler.JoinRoom)
	// 等待开始页
	http.HandleFunc("/room/", handler.LoadRoom)
	http.HandleFunc("/quitRoom/", handler.QuitRoom)
	http.HandleFunc("/start/", handler.Start)
	// 游戏中页

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Error:", err)
	}
}
