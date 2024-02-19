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
	_, roomIndex := findRoom(roomId)
	game := &cfg.Rooms[roomIndex]
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
			toggleNight(game, playerId)
		}
		time.Sleep(time.Millisecond * 50)
	}
}

func toggleNight(game *model.Room, playerId string) {
	cfg := model.GetConfig()
	cfgMutex.Lock()
	defer cfgMutex.Unlock()

	var msg string

	if playerId == game.Host {
		// Stage + 1
		game.State.Stage += 1
		// 日转夜 Day+1
		if !game.State.Night {
			game.State.Day = game.State.Day + 1
			msg = fmt.Sprintf("第%d天，入夜~", game.State.Day)
		} else {
			msg = fmt.Sprintf("第%d天，天亮了~", game.State.Day)
		}
		// 存入总日志
		game.Log += msg + "\n"
		// 存入个人日志，刷新的时候加载
		for i := range game.Players {
			game.Players[i].Log += msg + "\n"
		}
		// 日夜切换
		game.State.Night = !game.State.Night
		// 调整环节
		game.State.CastingStep = true
		game.State.CheckoutStep = false

		for i := range game.Players {
			// 调整玩家准备状态
			game.Players[i].Ready.Casted = false
			game.Players[i].Ready.Nominated = false
			game.Players[i].Ready.Voted = false
			// 让所有活人重新可以投票
			if !game.Players[i].State.Dead {
				game.Players[i].State.Vote = 1
			}
			// 管家无票
			if game.Players[i].Character == Butler {
				game.Players[i].State.Vote = 0
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
