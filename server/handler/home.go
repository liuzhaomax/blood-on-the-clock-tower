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

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway) {
				log.Println("Client disconnected:", err)
				return
			}
			log.Println("Read error Home:", err)
			return
		}

		var reqBody model.HomeReqBody
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
	cfg.HomeConnPool.Store(playerId, conn)
	// 发送房间列表给请求者
	marshalRooms, err := json.Marshal(cfg.Rooms)
	if err != nil {
		log.Println("JSON marshal error:", err)
		return
	}
	connVal, _ := cfg.HomeConnPool.LoadOrStore(playerId, conn)
	if err = connVal.(*websocket.Conn).WriteMessage(websocket.TextMessage, marshalRooms); err != nil {
		log.Println("Write error:", err)
		return
	}
}

func createRoom(room model.Room) {
	cfg := model.GetConfig()
	CfgMutex.Lock()
	defer CfgMutex.Unlock()

	room.Status = RoomWaitingToGo
	room.CreatedAt = time.Now().Format(time.RFC3339)
	cfg.Rooms = append(cfg.Rooms, room)
	// 发送房间列表给所有人
	marshalRooms, err := json.Marshal(cfg.Rooms)
	if err != nil {
		log.Println("JSON marshal error:", err)
		return
	}
	cfg.HomeConnPool.Range(func(id, conn any) bool {
		// 关闭创建房间者的连接
		if id == room.Players[0].Id {
			conn.(*websocket.Conn).Close()
			cfg.HomeConnPool.Delete(id)
			return true
		}
		if err = conn.(*websocket.Conn).WriteMessage(websocket.TextMessage, marshalRooms); err != nil {
			log.Println("Write error:", err)
			return false
		}
		return true
	})
}

func joinRoom(joinRoomPayload model.JoinRoomPayload) {
	cfg := model.GetConfig()
	CfgMutex.Lock()
	defer CfgMutex.Unlock()

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

	// 发送房间列表给home页所有人
	marshalRooms, err := json.Marshal(cfg.Rooms)
	if err != nil {
		log.Println("JSON marshal error:", err)
		return
	}
	cfg.HomeConnPool.Range(func(id, conn any) bool {
		// 关闭创建房间者的连接
		if id == joinRoomPayload.Player.Id {
			conn.(*websocket.Conn).Close()
			cfg.HomeConnPool.Delete(id)
			return true
		}
		if err = conn.(*websocket.Conn).WriteMessage(websocket.TextMessage, marshalRooms); err != nil {
			log.Println("Write error:", err)
			return false
		}
		return true
	})

	// 发送房间给所有人
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
}
