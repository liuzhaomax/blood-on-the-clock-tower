package handler

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/model"
	"log"
)

func reviewGame(room *model.Room, playerId string, conn *websocket.Conn) {
	cfg := model.GetConfig()
	CfgMutex.Lock()

	if cfg.RoomConnPool[room.Id] == nil {
		cfg.RoomConnPool[room.Id] = map[string]*websocket.Conn{}
	}
	cfg.RoomConnPool[room.Id][playerId] = conn

	CfgMutex.Unlock()

	// 把room发给请求者
	marshalRoom, err := json.Marshal(*room)
	if err != nil {
		log.Println("JSON marshal error:", err)
		return
	}
	if err := conn.WriteMessage(websocket.TextMessage, marshalRoom); err != nil {
		log.Println("Write error:", err)
		return
	}
	// 将room list 发给所有homeConn池里的人
	marshalRooms, err := json.Marshal(cfg.Rooms)
	if err != nil {
		log.Println("JSON marshal error:", err)
		return
	}
	for _, conn := range cfg.HomeConnPool {
		if err := conn.WriteMessage(websocket.TextMessage, marshalRooms); err != nil {
			log.Println("Write error:", err)
			return
		}
	}
}

func returnRoom(room *model.Room, playerId string) {
	cfg := model.GetConfig()
	CfgMutex.Lock()

	for i, player := range room.Players {
		if player.Id == playerId {
			room.Players[i].Waiting = true
			break
		}
	}

	CfgMutex.Unlock()

	// 把room发给请求者
	marshalRoom, err := json.Marshal(*room)
	if err != nil {
		log.Println("JSON marshal error:", err)
		return
	}
	if cfg.RoomConnPool[room.Id] == nil {
		return // 防空指针异常
	}
	if err := cfg.RoomConnPool[room.Id][playerId].WriteMessage(websocket.TextMessage, marshalRoom); err != nil {
		log.Println("Write error:", err)
		return
	}
}
