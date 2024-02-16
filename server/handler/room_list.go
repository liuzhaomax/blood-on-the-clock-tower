package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/model"
	"log"
	"net/http"
	"strings"
)

func ListRooms(w http.ResponseWriter, r *http.Request) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	messageType, _, err := conn.ReadMessage()
	if err != nil {
		if websocket.IsCloseError(err, websocket.CloseGoingAway) {
			log.Println("Client disconnected:", err)
			return
		}
		log.Println("Read error:", err)
		return
	}

	cfgMutex.Lock()
	defer cfgMutex.Unlock()
	cfg := model.GetConfig()
	marshalRooms, err := json.Marshal(cfg.Rooms)
	if err != nil {
		log.Println("JSON marshal error:", err)
		return
	}

	if err = conn.WriteMessage(messageType, marshalRooms); err != nil {
		log.Println("Write error:", err)
		return
	}
}

func CreateRoom(w http.ResponseWriter, r *http.Request) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	messageType, p, err := conn.ReadMessage()
	if err != nil {
		if websocket.IsCloseError(err, websocket.CloseGoingAway) {
			log.Println("Client disconnected:", err)
			return
		}
		log.Println("Read error:", err)
		return
	}

	var room model.Room
	if err = json.Unmarshal(p, &room); err != nil {
		log.Println("JSON unmarshal error:", err)
	}
	cfgMutex.Lock()
	defer cfgMutex.Unlock()
	cfg := model.GetConfig()
	room.State = "等待开始"

	// TODO 删除测试代码
	var player model.Player
	for i := 0; i < 13; i++ {
		player.Id = fmt.Sprintf("111111111%d", i)
		player.Name = fmt.Sprintf("机器人%d", i)
		room.Players = append(room.Players, player)
	}
	// TODO 测试代码结尾

	cfg.Rooms = append(cfg.Rooms, room)

	if err = conn.WriteMessage(messageType, p); err != nil {
		log.Println("Write error:", err)
		return
	}
}

func JoinRoom(w http.ResponseWriter, r *http.Request) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	messageType, p, err := conn.ReadMessage()
	if err != nil {
		if websocket.IsCloseError(err, websocket.CloseGoingAway) {
			log.Println("Client disconnected:", err)
			return
		}
		log.Println("Read error:", err)
		return
	}

	// 获取URL参数
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.NotFound(w, r)
		return
	}
	roomIdIn := parts[2]
	roomPasswordIn := parts[3]

	cfgMutex.Lock()
	defer cfgMutex.Unlock()
	cfg := model.GetConfig()
	room, roomIndex := findRoom(roomIdIn)

	if room.Password != roomPasswordIn {
		return
	}

	var playerIn model.Player
	if err = json.Unmarshal(p, &playerIn); err != nil {
		log.Println("JSON unmarshal error:", err)
	}

	var hasCurrentPlayer bool
	for i, player := range room.Players {
		if playerIn.Id == player.Id {
			cfg.Rooms[roomIndex].Players[i].Name = playerIn.Name
			hasCurrentPlayer = true
			break
		}
	}
	if !hasCurrentPlayer {
		cfg.Rooms[roomIndex].Players = append(room.Players, playerIn)
	}

	if err = conn.WriteMessage(messageType, p); err != nil {
		log.Println("Write error:", err)
		return
	}
}
