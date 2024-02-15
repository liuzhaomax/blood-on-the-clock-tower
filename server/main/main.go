package main

import (
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/handler"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/home", handler.ListRooms)
	http.HandleFunc("/room", handler.CreateRoom)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Error:", err)
	}
}
