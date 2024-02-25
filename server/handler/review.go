package handler

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/model"
	"log"
	"net/http"
	"strings"
)

func LoadReview(w http.ResponseWriter, r *http.Request) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	// 获取URL参数
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		http.NotFound(w, r)
		return
	}
	roomId := parts[2]

	cfg := model.GetConfig()
	cfgMutex.Lock()
	defer cfgMutex.Unlock()
	room, roomIndex := findRoom(roomId)

	room.Status = "复盘中"

	marshaledRoom, err := json.Marshal(cfg.Rooms[roomIndex])
	if err != nil {
		log.Println("JSON marshal error:", err)
		return
	}

	if err = conn.WriteMessage(websocket.TextMessage, marshaledRoom); err != nil {
		log.Println("Write error:", err)
		return
	}
}

func ReturnRoom(w http.ResponseWriter, r *http.Request) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	// 获取URL参数
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		http.NotFound(w, r)
		return
	}
	roomId := parts[2]

	cfgMutex.Lock()
	defer cfgMutex.Unlock()
	room, _ := findRoom(roomId)

	room.Status = "等待开始"
	room.Init = false
	room.Result = ""
	room.Log = ""
	for i, player := range room.Players {
		newPlayer := model.Player{}
		newPlayer.Id = player.Id
		newPlayer.Name = player.Name
		newPlayer.Index = player.Index
		room.Players[i] = newPlayer
	}
}
