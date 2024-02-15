package handler

import (
	"encoding/json"
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
	cfg := model.GetConfig()
	cfg.Rooms = append(cfg.Rooms, room)

	if err = conn.WriteMessage(messageType, p); err != nil {
		log.Println("Write error:", err)
		return
	}
}

func LoadRoom(w http.ResponseWriter, r *http.Request) {
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

	// 解析 URL，获取路径
	path := r.URL.Path
	// 提取 roomId 参数
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		http.NotFound(w, r)
		return
	}
	roomId := parts[2]
	room := findRoom(roomId)

	if string(p) == "quit_room" {
		playerId := parts[3]
		var newPlayers []model.Player
		for _, player := range room.Players {
			if playerId != player.Id {
				newPlayers = append(newPlayers, player)
			}
		}
		room.Players = newPlayers
		if len(room.Players) == 0 {
			cfg := model.GetConfig()
			var newRooms []model.Room
			for _, roomm := range cfg.Rooms {
				if roomId != roomm.Id {
					newRooms = append(newRooms, roomm)
				}
			}
			cfg.Rooms = newRooms
		}
	} else {
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
}

func findRoom(roomId string) *model.Room {
	cfg := model.GetConfig()
	for _, room := range cfg.Rooms {
		if room.Id == roomId {
			return &room
		}
	}
	return nil
}
