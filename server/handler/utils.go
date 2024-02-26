package handler

import (
	"github.com/gorilla/websocket"
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/model"
	"math/rand"
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
			return &cfg.Rooms[i], i
		}
	}
	return nil, 0
}

func Shuffle(strSlice []string) []string {
	// 使用 Fisher-Yates 算法随机排列切片中的元素
	for i := len(strSlice) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		strSlice[i], strSlice[j] = strSlice[j], strSlice[i]
	}
	return strSlice
}

func Contains(s []string, e string) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
