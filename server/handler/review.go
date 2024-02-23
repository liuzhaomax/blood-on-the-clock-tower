package handler

import (
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/model"
	"log"
	"net/http"
	"strings"
)

func LoadReview(w http.ResponseWriter, r *http.Request) {
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

	cfgMutex.Lock()
	defer cfgMutex.Unlock()
	room, _ := findRoom(roomId)

	room.Status = "复盘中"
}

func ReturnRoom(w http.ResponseWriter, r *http.Request) {
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

	cfg := model.GetConfig()
	cfgMutex.Lock()
	defer cfgMutex.Unlock()
	room, _ := findRoom(roomId)

	room.Status = "等待开始"
	room.Init = false

	// 关闭房间内所有玩家的gaming长连接
	for id := range cfg.ConnPool {
		for _, player := range room.Players {
			if id == player.Id {
				err := cfg.ConnPool[id].Close()
				if err != nil {
					log.Println(err)
					return
				}
			}
		}
	}
}
