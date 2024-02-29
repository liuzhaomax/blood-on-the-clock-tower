package handler

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/model"
	"log"
	"net/http"
	"strings"
	"time"
)

func LoadRoom(w http.ResponseWriter, r *http.Request) {
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
	room, _ := findRoom(roomId)

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

		var reqBody model.RoomReqBody
		if err = json.Unmarshal(p, &reqBody); err != nil {
			log.Println("JSON unmarshal error:", err)
		}

		switch reqBody.Action {
		case "list_players":
			listPlayers(room, reqBody.Payload, conn)
		case "quit_room":
			quitRoom(room, reqBody.Payload)
		case "start_game":
			startGame(room)
		}

		time.Sleep(time.Millisecond * 50)
	}
}

func listPlayers(room *model.Room, playerId string, conn *websocket.Conn) {
	// 防止前端循环请求出现在quit之后
	if room == nil {
		return
	}
	cfg := model.GetConfig()
	cfgMutex.Lock()
	if cfg.RoomConnPool[room.Id] == nil {
		cfg.RoomConnPool[room.Id] = map[string]*websocket.Conn{}
	}
	cfg.RoomConnPool[room.Id][playerId] = conn
	cfgMutex.Unlock()
	// 发送房间列表给请求者
	marshalRoom, err := json.Marshal(*room)
	if err != nil {
		log.Println("JSON marshal error:", err)
		return
	}
	if err = cfg.RoomConnPool[room.Id][playerId].WriteMessage(websocket.TextMessage, marshalRoom); err != nil {
		log.Println("Write error:", err)
		return
	}
}

func quitRoom(room *model.Room, playerId string) {
	cfg := model.GetConfig()
	cfgMutex.Lock()
	defer cfgMutex.Unlock()

	// 删除退出者
	var newPlayers []model.Player
	for _, player := range room.Players {
		if playerId != player.Id {
			newPlayers = append(newPlayers, player)
		}
	}
	room.Players = newPlayers
	// 所有人都退出了
	if len(room.Players) == 0 {
		var newRooms []model.Room
		for _, roomm := range cfg.Rooms {
			if room.Id != roomm.Id {
				newRooms = append(newRooms, roomm)
			}
		}
		cfg.Rooms = newRooms
	}

	// 发送房间给所有人
	marshalRoom, err := json.Marshal(*room)
	if err != nil {
		log.Println("JSON marshal error:", err)
		return
	}
	for id, conn := range cfg.RoomConnPool[room.Id] {
		// 关闭退出房间者的连接
		if id == playerId {
			conn.Close()
			delete(cfg.RoomConnPool[room.Id], id)
			continue
		}
		if err := conn.WriteMessage(websocket.TextMessage, marshalRoom); err != nil {
			log.Println("Write error:", err)
			return
		}
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

func startGame(room *model.Room) {
	cfg := model.GetConfig()
	cfgMutex.Lock()
	defer cfgMutex.Unlock()

	goodToStart := true
	for _, player := range room.Players {
		goodToStart = goodToStart && player.Waiting
	}
	if goodToStart {
		room.Status = "游戏中"
		room.Init = false
		room.Result = ""
		room.Log = ""
		room.CastPool = map[string][]string{}
		room.VotePool = map[string]string{}
		room.State = model.GameState{}
	}
	// 发送房间给所有人
	marshalRoom, err := json.Marshal(*room)
	if err != nil {
		log.Println("JSON marshal error:", err)
		return
	}
	for _, conn := range cfg.RoomConnPool[room.Id] {
		if err := conn.WriteMessage(websocket.TextMessage, marshalRoom); err != nil {
			log.Println("Write error:", err)
			return
		}
		conn.Close()
	}
	delete(cfg.RoomConnPool, room.Id)
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
