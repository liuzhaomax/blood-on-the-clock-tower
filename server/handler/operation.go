package handler

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/model"
	"log"
	"net/http"
	"strings"
	"time"
)

func Gaming(w http.ResponseWriter, r *http.Request) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	// 获取URL参数
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.NotFound(w, r)
		return
	}
	roomId := parts[2]
	playerId := parts[3]

	// 推入连接池
	cfg := model.GetConfig()
	cfg.ConnPool[playerId] = conn
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

		action := string(p)
		switch action {
		case "toggle_night":
			toggleNight(roomId, playerId)
		}
		time.Sleep(time.Millisecond * 50)
	}
}

func toggleNight(roomId string, playerId string) {
	cfgMutex.Lock()
	defer cfgMutex.Unlock()
	cfg := model.GetConfig()
	_, roomIndex := findRoom(roomId)

	var msg string

	if playerId == cfg.Rooms[roomIndex].Host {
		if !cfg.Rooms[roomIndex].Night {
			cfg.Rooms[roomIndex].Day = cfg.Rooms[roomIndex].Day + 1
			msg = fmt.Sprintf("第%d天，入夜~", cfg.Rooms[roomIndex].Day)
		} else {
			msg = fmt.Sprintf("第%d天，天亮了~", cfg.Rooms[roomIndex].Day)
		}
		// 存入总日志
		cfg.Rooms[roomIndex].Log += msg + "\n"
		// 存入个人日志，刷新的时候加载
		for i := range cfg.Rooms[roomIndex].Players {
			cfg.Rooms[roomIndex].Players[i].Log += msg + "\n"
		}

		cfg.Rooms[roomIndex].Night = !cfg.Rooms[roomIndex].Night

		// 让所有活人重新可以投票
		for i := range cfg.Rooms[roomIndex].Players {
			if !cfg.Rooms[roomIndex].Players[i].Status.Dead {
				cfg.Rooms[roomIndex].Players[i].Status.Vote = true
			}
		}
	}

	// 将日夜切换日志群发
	for _, conn := range cfg.ConnPool {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			log.Println("Write error:", err)
			return
		}
	}
}
