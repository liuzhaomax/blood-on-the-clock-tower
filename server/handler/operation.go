package handler

import (
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/model"
	"log"
	"net/http"
	"strings"
)

func ToggleNight(w http.ResponseWriter, r *http.Request) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	// 获取URL参数
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.NotFound(w, r)
		return
	}
	roomId := parts[3]

	cfgMutex.Lock()
	defer cfgMutex.Unlock()
	cfg := model.GetConfig()
	_, roomIndex := findRoom(roomId)

	if !cfg.Rooms[roomIndex].Night {
		cfg.Rooms[roomIndex].Day = cfg.Rooms[roomIndex].Day + 1
	}
	cfg.Rooms[roomIndex].Night = !cfg.Rooms[roomIndex].Night
}
