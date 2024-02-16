package handler

import (
	"github.com/gorilla/websocket"
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/model"
	"net/http"
	"sync"
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var cfgMutex sync.Mutex

func findRoom(roomId string) (*model.Room, int) {
	cfg := model.GetConfig()
	for i, room := range cfg.Rooms {
		if room.Id == roomId {
			return &room, i
		}
	}
	return nil, 0
}
