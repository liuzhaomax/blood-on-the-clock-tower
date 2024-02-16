package handler

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/model"
	"log"
	"net/http"
	"strings"
)

func LoadRoom(w http.ResponseWriter, r *http.Request) {
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

	// 防止前端循环请求出现在quit之后
	if room == nil {
		return
	}

	marshalRoom, err := json.Marshal(room)
	if err != nil {
		log.Println("JSON marshal error:", err)
		return
	}

	if err = conn.WriteMessage(messageType, marshalRoom); err != nil {
		log.Println("Write error:", err)
		return
	}
}

func QuitRoom(w http.ResponseWriter, r *http.Request) {
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
	if len(parts) < 3 {
		http.NotFound(w, r)
		return
	}
	roomId := parts[2]

	var playerIn model.Player
	if err = json.Unmarshal(p, &playerIn); err != nil {
		log.Println("JSON unmarshal error:", err)
	}

	cfgMutex.Lock()
	defer cfgMutex.Unlock()
	cfg := model.GetConfig()
	room, roomIndex := findRoom(roomId)

	var newPlayers []model.Player
	for _, player := range room.Players {
		if playerIn.Id != player.Id {
			newPlayers = append(newPlayers, player)
		}
	}
	cfg.Rooms[roomIndex].Players = newPlayers
	if len(cfg.Rooms[roomIndex].Players) == 0 {
		var newRooms []model.Room
		for _, roomm := range cfg.Rooms {
			if roomId != roomm.Id {
				newRooms = append(newRooms, roomm)
			}
		}
		cfg.Rooms = newRooms
	}

	if err = conn.WriteMessage(messageType, p); err != nil {
		log.Println("Write error:", err)
		return
	}
}
