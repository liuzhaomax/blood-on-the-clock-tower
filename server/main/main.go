package main

import (
	"encoding/json"
	"flag"
	"github.com/gorilla/websocket"
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/handler"
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/model"
	"log"
	"net/http"
	"time"
)

func main() {
	log.Println("服务启动")
	// 房间列表页
	http.HandleFunc("/home", handler.LoadHome)
	// 等待开始页
	http.HandleFunc("/room/", handler.LoadRoom)
	// 游戏中页
	http.HandleFunc("/game/", handler.LoadGame)

	// 检测是否5小时没有操作，是就销毁房间
	go func() {
		for {
			cfg := model.GetConfig()
			longerRoomIndex := []int{}
			for i, room := range cfg.Rooms {
				longer, err := handler.HasPassedGivenHours(room.CreatedAt, 5)
				if err != nil {
					log.Println("出现时间解析错误：", err)
					continue
				}
				if longer {
					longerRoomIndex = append(longerRoomIndex, i)
				}
			}
			handler.CfgMutex.Lock()
			for _, roomIndex := range longerRoomIndex {
				cfg.Rooms = handler.RemoveElement(cfg.Rooms, roomIndex)
			}
			handler.CfgMutex.Unlock()
			// 给所有home页的人发送房间列表
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

			time.Sleep(time.Hour * 4)
		}
	}()

	var mode string
	flag.StringVar(&mode, "mode", "dev", "dev or prod")
	flag.Parse()
	switch mode {
	case "dev":
		err := http.ListenAndServe(":62121", nil)
		if err != nil {
			log.Fatal("Error:", err)
		}
	case "prod":
		err := http.ListenAndServeTLS(":62121", "tls/bloodws.liuzhaomax.cn.pem", "tls/bloodws.liuzhaomax.cn.key", nil)
		if err != nil {
			log.Fatal("Error:", err)
		}
	}
}
