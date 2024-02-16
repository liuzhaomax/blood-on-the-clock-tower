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
	http.HandleFunc("/startGame/", handler.StartGame)
	// 游戏中页
	http.HandleFunc("/game/", handler.LoadGame)
	// 游戏结算页
	// http.HandleFunc("/checkout/", handler.LoadCheckout)
	// http.HandleFunc("/return/", handler.ReturnRoom)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Error:", err)
	}
}