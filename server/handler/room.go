package handler

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/model"
	"log"
	"net/http"
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
