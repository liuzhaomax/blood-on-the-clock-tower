package handler

import (
	"encoding/json"
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
	// 技能施放池，存储所有施放技能人，当前阶段施放的技能作用目标
	game.CastPool = map[string][]string{}

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

		var actionReq model.ActionReq
		if err = json.Unmarshal(p, &actionReq); err != nil {
			log.Println("JSON unmarshal error:", err)
		}

		switch actionReq.Action {
		case "toggle_night":
			toggleNight(game, playerId)
		case "cast":
			cast(game, playerId, actionReq.Targets)
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

		for i := range game.Players {
			// 调整玩家准备状态
			game.Players[i].Ready.Casted = false
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

func cast(game *model.Room, playerId string, targets []string) {
	cfg := model.GetConfig()
	cfgMutex.Lock()
	defer cfgMutex.Unlock()

	var msgPlayer = "您"
	var msgAll string
	for i, player := range game.Players {
		if player.Id == playerId {
			msgAll += player.Name
			switch player.Character {
			case Poisoner:
				for _, player := range game.Players {
					if targets[0] == player.Id {
						info := fmt.Sprintf("对 [%s] 进行了下毒！", player.Name)
						msgPlayer += info
						msgAll += info
						break
					}
				}
			case FortuneTeller:
				for _, player := range game.Players {
					if targets[0] == player.Id {
						info := "对 [" + player.Name + "]、["
						msgPlayer += info
						msgAll += info
						break
					}
				}
				for _, player := range game.Players {
					if targets[1] == player.Id {
						info := player.Name + "] 进行了占卜！"
						msgPlayer += info
						msgAll += info
						break
					}
				}
			case Butler:
				for _, player := range game.Players {
					if targets[0] == player.Id {
						info := fmt.Sprintf(" 对 [%s] 进行了认主！", player.Name)
						msgPlayer += info
						msgAll += info
						break
					}
				}
			case Monk:
				for _, player := range game.Players {
					if targets[0] == player.Id {
						info := fmt.Sprintf(" 对 [%s] 进行了守护！", player.Name)
						msgPlayer += info
						msgAll += info
						break
					}
				}
			case Imp:
				for _, player := range game.Players {
					if targets[0] == player.Id {
						info := fmt.Sprintf(" 对 [%s] 进行了杀害！", player.Name)
						msgPlayer += info
						msgAll += info
						break
					}
				}
			case Slayer:
				for _, player := range game.Players {
					if targets[0] == player.Id {
						info := fmt.Sprintf(" 对 [%s] 进行了枪毙！", player.Name)
						msgPlayer += info
						msgAll += info
						break
					}
				}
			}
			game.Players[i].Ready.Casted = true
			break
		}
	}
	game.CastPool[playerId] = targets
	for i, player := range game.Players {
		if player.Id == playerId {
			game.Players[i].Log += msgPlayer
			break
		}
	}
	game.Log += msgAll + "\n"

	// 发送日志
	for _, conn := range cfg.ConnPool {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msgPlayer)); err != nil {
			log.Println("Write error:", err)
			return
		}
	}
}
