package handler

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/model"
	"log"
	"net/http"
	"time"
)

func LoadHome(w http.ResponseWriter, r *http.Request) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway) {
				log.Println("Client disconnected:", err)
				return
			}
			log.Println("Read error:", err)
			return
		}

		var reqBody model.ListRoomsReqBody
		if err = json.Unmarshal(p, &reqBody); err != nil {
			log.Println("JSON unmarshal error:", err)
		}

		switch reqBody.Action {
		case "list_rooms":
			var reqBody model.ListRoomsReqBody
			if err = json.Unmarshal(p, &reqBody); err != nil {
				log.Println("JSON unmarshal error:", err)
			}
			listRooms(reqBody.Payload, conn)
		case "create_room":
			var reqBody model.CreateRoomReqBody
			if err = json.Unmarshal(p, &reqBody); err != nil {
				log.Println("JSON unmarshal error:", err)
			}
			createRoom(reqBody.Payload)
		case "join_room":
			var reqBody model.JoinRoomReqBody
			if err = json.Unmarshal(p, &reqBody); err != nil {
				log.Println("JSON unmarshal error:", err)
			}
			joinRoom(reqBody.Payload)
		}

		time.Sleep(time.Millisecond * 50)
	}
}

func listRooms(playerId string, conn *websocket.Conn) {
	cfg := model.GetConfig()
	cfgMutex.Lock()
	cfg.HomeConnPool[playerId] = conn
	cfgMutex.Unlock()
	// 发送房间列表给请求者
	marshalRooms, err := json.Marshal(cfg.Rooms)
	if err != nil {
		log.Println("JSON marshal error:", err)
		return
	}
	if err = cfg.HomeConnPool[playerId].WriteMessage(websocket.TextMessage, marshalRooms); err != nil {
		log.Println("Write error:", err)
		return
	}
}

func createRoom(room model.Room) {
	cfg := model.GetConfig()
	cfgMutex.Lock()
	room.Status = "等待开始"
	cfg.Rooms = append(cfg.Rooms, room)
	cfgMutex.Unlock()
	// 发送房间列表给所有人
	marshalRooms, err := json.Marshal(cfg.Rooms)
	if err != nil {
		log.Println("JSON marshal error:", err)
		return
	}
	for id, conn := range cfg.HomeConnPool {
		// 关闭创建房间者的连接
		if id == room.Players[0].Id {
			conn.Close()
			delete(cfg.HomeConnPool, id)
			continue
		}
		if err = conn.WriteMessage(websocket.TextMessage, marshalRooms); err != nil {
			log.Println("Write error:", err)
			return
		}
	}
}

func joinRoom(joinRoomPayload model.JoinRoomPayload) {
	cfg := model.GetConfig()
	cfgMutex.Lock()

	room, roomIndex := findRoom(joinRoomPayload.Room.Id)

	if room.Password != joinRoomPayload.Room.Password {
		return
	}

	var hasCurrentPlayer bool
	for i, player := range room.Players {
		if joinRoomPayload.Player.Id == player.Id {
			cfg.Rooms[roomIndex].Players[i].Name = joinRoomPayload.Player.Name
			hasCurrentPlayer = true
			break
		}
	}
	if !hasCurrentPlayer {
		cfg.Rooms[roomIndex].Players = append(room.Players, joinRoomPayload.Player)
	}

	cfgMutex.Unlock()

	// 发送房间列表给所有人
	marshalRooms, err := json.Marshal(cfg.Rooms)
	if err != nil {
		log.Println("JSON marshal error:", err)
		return
	}
	for id, conn := range cfg.HomeConnPool {
		// 关闭加入房间者的连接
		if id == joinRoomPayload.Player.Id {
			conn.Close()
			delete(cfg.HomeConnPool, id)
			continue
		}
		if err = conn.WriteMessage(websocket.TextMessage, marshalRooms); err != nil {
			log.Println("Write error:", err)
			return
		}
	}
}
