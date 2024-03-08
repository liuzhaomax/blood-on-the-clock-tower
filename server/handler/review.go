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
	defer CfgMutex.Unlock()
	if room == nil {
		return
	}
	room.GameConnPool.Store(playerId, conn)

	// 把room发给请求者
	marshalRoom, err := json.Marshal(*room)
	if err != nil {
		log.Println("JSON marshal error:", err)
		return
	}
	connVal, _ := room.GameConnPool.LoadOrStore(playerId, conn)
	if err = connVal.(*websocket.Conn).WriteMessage(websocket.TextMessage, marshalRoom); err != nil {
		log.Println("Write error:", err)
		return
	}
	// 将room list 发给所有homeConn池里的人
	marshalRooms, err := json.Marshal(cfg.Rooms)
	if err != nil {
		log.Println("JSON marshal error:", err)
		return
	}
	cfg.HomeConnPool.Range(func(id, conn any) bool {
		if err = conn.(*websocket.Conn).WriteMessage(websocket.TextMessage, marshalRooms); err != nil {
			log.Println("Write error:", err)
			return false
		}
		return true
	})
}

func returnRoom(room *model.Room, playerId string) {
	CfgMutex.Lock()
	defer CfgMutex.Unlock()

	for i, player := range room.Players {
		if player.Id == playerId {
			room.Players[i].Waiting = true
			break
		}
	}

	roomReady := true
	for _, player := range room.Players {
		roomReady = roomReady && player.Waiting
	}
	if roomReady {
		room.Status = RoomWaitingToGo
	}

	// 把room发给房间内所有人
	marshalRoom, err := json.Marshal(*room)
	if err != nil {
		log.Println("JSON marshal error:", err)
		return
	}
	room.GameConnPool.Range(func(id, conn any) bool {
		if err = conn.(*websocket.Conn).WriteMessage(websocket.TextMessage, marshalRoom); err != nil {
			log.Println("Write error:", err)
			return false
		}
		return true
	})
	// 把roomList发给房间外所有人
	cfg := model.GetConfig()
	marshalRooms, err := json.Marshal(cfg.Rooms)
	if err != nil {
		log.Println("JSON marshal error:", err)
		return
	}
	cfg.HomeConnPool.Range(func(id, conn any) bool {
		if err = conn.(*websocket.Conn).WriteMessage(websocket.TextMessage, marshalRooms); err != nil {
			log.Println("Write error:", err)
			return false
		}
		return true
	})
}
