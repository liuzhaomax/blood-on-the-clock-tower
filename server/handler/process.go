package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/model"
	"log"
	"math/rand"
	"net/http"
	"reflect"
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
			if playerId == game.Host {
				toggleNight(game)
			}
		case "cast":
			cast(game, playerId, actionReq.Targets)
		case "checkout":
			if playerId == game.Host {
				checkout(game, playerId)
			}
		}
		time.Sleep(time.Millisecond * 50)
	}
}

func toggleNight(game *model.Room) {
	cfg := model.GetConfig()
	cfgMutex.Lock()
	defer cfgMutex.Unlock()

	var msg string

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
		// 让所有活人重新可以投票
		if !game.Players[i].State.Dead {
			game.Players[i].State.Vote = 1
		}
		// 管家无票
		if game.Players[i].Character == Butler {
			game.Players[i].State.Vote = 0
		}
		// 调整玩家施放技能的准备状态
		game.Players[i].Ready.Casted = true
		if !game.Players[i].State.Dead {
			switch game.Players[i].Character {
			case Poisoner:
				if game.State.Stage%2 == 1 {
					game.Players[i].Ready.Casted = false
				}
			case FortuneTeller:
				if game.State.Stage%2 == 1 {
					game.Players[i].Ready.Casted = false
				}
			case Butler:
				if game.State.Stage%2 == 1 {
					game.Players[i].Ready.Casted = false
				}
			case Monk:
				if game.State.Stage%2 == 1 && game.State.Stage != 1 {
					game.Players[i].Ready.Casted = false
				}
			case Imp:
				if game.State.Stage%2 == 1 && game.State.Stage != 1 {
					game.Players[i].Ready.Casted = false
				}
			case Slayer:
				if game.State.Stage%2 == 0 && game.Players[i].State.Bullet {
					game.Players[i].Ready.Casted = false
				}
			default:
				game.Players[i].Ready.Casted = true
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
	var msgAll = ""
	for i, player := range game.Players {
		if player.Id == playerId {
			msgAll += fmt.Sprintf("[%s] ", player.Name)
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
			game.Players[i].Log += msgPlayer + "\n"
			break
		}
	}
	game.Log += msgAll + "\n"

	// 发送日志
	for id, conn := range cfg.ConnPool {
		if id == playerId {
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msgPlayer)); err != nil {
				log.Println("Write error:", err)
				return
			}
			break
		}
	}
}

func checkout(game *model.Room, playerId string) {
	cfg := model.GetConfig()
	cfgMutex.Lock()
	defer cfgMutex.Unlock()

	var msgPlayer = "您"
	var msgAll = ""

	castPoolObj := map[model.Player][]int{}
	for fromPlayerId, toPlayerIdSlice := range game.CastPool {
		for _, player := range game.Players {
			if player.Id == fromPlayerId {
				castPoolObj[player] = []int{}
				break
			}
		}
		for _, toPlayerId := range toPlayerIdSlice {
			for _, player := range game.Players {
				if player.Id == toPlayerId {
					castPoolObj[player] = append(castPoolObj[player], player.Index)
					break
				}
			}
		}
	}
	// 判断下毒
	for fromPlayer, toPlayerIndexSlice := range castPoolObj {
		if fromPlayer.Character == Poisoner {
			game.Players[toPlayerIndexSlice[0]].State.Poisoned = true
			break
		}
	}
	// 结算第一夜信息
	if game.State.Stage%2 == 1 {
		// if game.State.Stage == 1 {
		for i, player := range game.Players {
			msgPlayer = "您"
			msgAll = ""
			switch player.Character {
			// 给洗衣妇提供信息
			case Washerwoman:
				var realFolk model.Player
				var other model.Player
				if !player.State.Drunk && !player.State.Poisoned {
					// 生成随机信息
					for {
						randInt := rand.Intn(len(game.Players))
						if game.Players[randInt].CharacterType == Townsfolk && game.Players[randInt].Id != playerId {
							realFolk = game.Players[randInt]
							break
						}
					}
					for {
						randInt := rand.Intn(len(game.Players))
						if game.Players[randInt].Id != realFolk.Id && game.Players[randInt].Id != playerId {
							other = game.Players[randInt]
							break
						}
					}
				} else {
					// 生成伪信息
					for {
						randInt := rand.Intn(len(game.Players))
						if game.Players[randInt].Id != playerId {
							realFolk = game.Players[randInt]
							break
						}
					}
					for {
						randInt := rand.Intn(len(game.Players))
						if game.Players[randInt].Id != realFolk.Id && game.Players[randInt].Id != playerId {
							other = game.Players[randInt]
							break
						}
					}
				}
				// 拼接日志
				msgAll += fmt.Sprintf("[%s] ", player.Name)
				var info string
				randInt := rand.Intn(2)
				if randInt == 0 {
					info = fmt.Sprintf("发现 [%s] 和 [%s] 其中一个是 {%s}", realFolk.Name, other.Name, realFolk.Character)
				} else {
					info = fmt.Sprintf("发现 [%s] 和 [%s] 其中一个是 {%s}", other.Name, realFolk.Name, realFolk.Character)
				}
				msgPlayer += info
				msgAll += info
				game.Players[i].Log += msgPlayer + "\n"
				game.Log += msgAll + "\n"
				// 发送日志
				for id, conn := range cfg.ConnPool {
					if id == player.Id {
						if err := conn.WriteMessage(websocket.TextMessage, []byte(msgPlayer)); err != nil {
							log.Println("Write error:", err)
							return
						}
						break
					}
				}
			// 给图书管理员提供信息
			case Librarian:
				var realOutsider model.Player
				var other model.Player
				if !player.State.Drunk && !player.State.Poisoned {
					var hasOutsider bool
					for _, player := range game.Players {
						if player.CharacterType == Outsiders {
							hasOutsider = true
							break
						}
					}
					if hasOutsider {
						// 生成随机信息
						for {
							randInt := rand.Intn(len(game.Players))
							if game.Players[randInt].CharacterType == Outsiders && game.Players[randInt].Id != playerId {
								realOutsider = game.Players[randInt]
								break
							}
						}
						for {
							randInt := rand.Intn(len(game.Players))
							if game.Players[randInt].Id != realOutsider.Id && game.Players[randInt].Id != playerId {
								other = game.Players[randInt]
								break
							}
						}
					}
				} else {
					// 生成伪信息
					var hasOutsider bool
					randInt := rand.Intn(len(game.Players))
					if game.Players[randInt].Id != playerId {
						realOutsider = game.Players[randInt]
						hasOutsider = true
					}
					if hasOutsider {
						for {
							randInt := rand.Intn(len(game.Players))
							if game.Players[randInt].Id != realOutsider.Id && game.Players[randInt].Id != playerId {
								other = game.Players[randInt]
								break
							}
						}
					}
				}
				// 拼接日志
				msgAll += fmt.Sprintf("[%s] ", player.Name)
				var info string
				if reflect.ValueOf(realOutsider).IsZero() {
					info = "发现本局 {没有外来者}"
				} else {
					randInt := rand.Intn(2)
					if randInt == 0 {
						info = fmt.Sprintf("发现 [%s] 和 [%s] 其中一个是 {%s}", realOutsider.Name, other.Name, realOutsider.Character)
					} else {
						info = fmt.Sprintf("发现 [%s] 和 [%s] 其中一个是 {%s}", other.Name, realOutsider.Name, realOutsider.Character)
					}
				}
				msgPlayer += info
				msgAll += info
				game.Players[i].Log += msgPlayer + "\n"
				game.Log += msgAll + "\n"
				// 发送日志
				for id, conn := range cfg.ConnPool {
					if id == player.Id {
						if err := conn.WriteMessage(websocket.TextMessage, []byte(msgPlayer)); err != nil {
							log.Println("Write error:", err)
							return
						}
						break
					}
				}
			// 给调查员提供信息
			case Investigator:
				var realMinion model.Player
				var other model.Player
				if !player.State.Drunk && !player.State.Poisoned {
					// 生成随机信息
					for {
						randInt := rand.Intn(len(game.Players))
						if game.Players[randInt].CharacterType == Minions && game.Players[randInt].Id != playerId {
							realMinion = game.Players[randInt]
							break
						}
					}
					for {
						randInt := rand.Intn(len(game.Players))
						if game.Players[randInt].Id != realMinion.Id && game.Players[randInt].Id != playerId {
							other = game.Players[randInt]
							break
						}
					}
				} else {
					// 生成伪信息
					for {
						randInt := rand.Intn(len(game.Players))
						if game.Players[randInt].Id != playerId {
							realMinion = game.Players[randInt]
							break
						}
					}
					for {
						randInt := rand.Intn(len(game.Players))
						if game.Players[randInt].Id != realMinion.Id && game.Players[randInt].Id != playerId {
							other = game.Players[randInt]
							break
						}
					}
				}
				// 拼接日志
				msgAll += fmt.Sprintf("[%s] ", player.Name)
				var info string
				randInt := rand.Intn(2)
				if randInt == 0 {
					info = fmt.Sprintf("发现 [%s] 和 [%s] 其中一个是 {%s}", realMinion.Name, other.Name, realMinion.Character)
				} else {
					info = fmt.Sprintf("发现 [%s] 和 [%s] 其中一个是 {%s}", other.Name, realMinion.Name, realMinion.Character)
				}
				msgPlayer += info
				msgAll += info
				game.Players[i].Log += msgPlayer + "\n"
				game.Log += msgAll + "\n"
				// 发送日志
				for id, conn := range cfg.ConnPool {
					if id == player.Id {
						if err := conn.WriteMessage(websocket.TextMessage, []byte(msgPlayer)); err != nil {
							log.Println("Write error:", err)
							return
						}
						break
					}
				}
			// 给厨师提供信息
			case Chef:
				connected := 0 // 记录连座数
				if !player.State.Drunk && !player.State.Poisoned {
					// 生成连座信息
					meetEvil := false
					meetEvilAgain := false
					for i, player := range game.Players {
						if player.State.Evil {
							if meetEvil {
								meetEvilAgain = true
							}
							if meetEvilAgain {
								connected += 1
								meetEvilAgain = false
							}
							if i == len(game.Players)-1 && game.Players[0].State.Evil {
								connected += 1
								break
							}
							meetEvil = true
						} else {
							meetEvil = false
						}
					}
				} else {
					// 生成伪信息
					evilQuantity := 0
					for _, player := range game.Players {
						if player.State.Evil {
							evilQuantity += 1
						}
					}
					randInt := rand.Intn(evilQuantity + 1)
					connected = randInt
				}
				// 拼接日志
				msgAll += fmt.Sprintf("[%s] ", player.Name)
				info := fmt.Sprintf("互为邻座的邪恶玩家有 {%d} 对", connected)
				msgPlayer += info
				msgAll += info
				game.Players[i].Log += msgPlayer + "\n"
				game.Log += msgAll + "\n"
				// 发送日志
				for id, conn := range cfg.ConnPool {
					if id == player.Id {
						if err := conn.WriteMessage(websocket.TextMessage, []byte(msgPlayer)); err != nil {
							log.Println("Write error:", err)
							return
						}
						break
					}
				}
			}
		}
	}
	// 判断占卜师
}
