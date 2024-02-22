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
	"sync"
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

	var mux *sync.Mutex

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

		// 被处决者
		executed := &model.Player{}

		switch actionReq.Action {
		case "toggle_night":
			if playerId == game.Host {
				toggleNight(mux, game)
			}
		case "cast":
			cast(mux, game, playerId, actionReq.Targets)
		case "checkout_night":
			if playerId == game.Host {
				checkoutNight(mux, game, playerId, executed)
			}
		}
		time.Sleep(time.Millisecond * 50)
	}
}

func toggleNight(mux *sync.Mutex, game *model.Room) {
	cfg := model.GetConfig()
	mux.Lock()
	defer mux.Unlock()

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
		// 让所有活人重新可以投票，夜转日结算，没投票还有票，
		// TODO 日转夜结算，被处决后要将vote赋为true
		if !game.Players[i].State.Dead {
			game.Players[i].State.Vote = 1
		}
		// 管家无票
		if game.Players[i].Character == Butler {
			game.Players[i].State.Vote = 0
		}
		// 调整玩家施放技能的准备状态
		game.Players[i].Ready.Casted = true
		game.CastPool = map[string][]string{}
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
		// 让所有人的僧侣守护状态消失
		game.Players[i].State.Protected = false
	}

	// 将日夜切换日志群发
	for _, conn := range cfg.ConnPool {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			log.Println("Write error:", err)
			return
		}
	}
}

func cast(mux *sync.Mutex, game *model.Room, playerId string, targets []string) {
	cfg := model.GetConfig()
	mux.Lock()
	defer mux.Unlock()

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
			case Ravenkeeper:
				for _, player := range game.Players {
					if targets[0] == player.Id {
						info := fmt.Sprintf(" 对 [%s] 进行了反向通灵！", player.Name)
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

func checkoutNight(mux *sync.Mutex, game *model.Room, playerId string, executed *model.Player) {
	cfg := model.GetConfig()
	mux.Lock()
	defer mux.Unlock()

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
	if game.State.Stage == 1 {
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
				info := fmt.Sprintf("发现互为邻座的邪恶玩家有 {%d} 对", connected)
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
			// 给共情者提供信息
			case Empath:
				evilQuantity := 0 // 记录左右邪恶玩家数量
				if !player.State.Drunk && !player.State.Poisoned {
					// 生成连座信息
					if player.Index == 0 {
						if game.Players[len(game.Players)-1].State.Evil {
							evilQuantity += 1
						}
						if game.Players[player.Index+1].State.Evil {
							evilQuantity += 1
						}
					} else if player.Index == len(game.Players)-1 {
						if game.Players[player.Index-1].State.Evil {
							evilQuantity += 1
						}
						if game.Players[0].State.Evil {
							evilQuantity += 1
						}
					} else {
						if game.Players[player.Index-1].State.Evil {
							evilQuantity += 1
						}
						if game.Players[player.Index+1].State.Evil {
							evilQuantity += 1
						}
					}
				} else {
					// 生成伪信息
					randInt := rand.Intn(3)
					evilQuantity = randInt
				}
				// 拼接日志
				msgAll += fmt.Sprintf("[%s] ", player.Name)
				info := fmt.Sprintf("发现与您邻座的邪恶玩家有 {%d} 个", evilQuantity)
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
			// 给间谍提供信息
			case Spy:
				// 拼接日志
				msgAll += fmt.Sprintf("[%s] 是间谍，知晓所有身份", player.Name)
				var info string
				for _, player := range game.Players {
					info += fmt.Sprintf("知晓所有身份：\n玩家 [%s] 的身份是 {%s}\n", player.Name, player.Character)
				}
				msgPlayer += info
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
	// 判断守护
	for fromPlayer, toPlayerIndexSlice := range castPoolObj {
		if fromPlayer.Character == Monk && !fromPlayer.State.Poisoned && !fromPlayer.State.Dead && !fromPlayer.State.Drunk {
			game.Players[toPlayerIndexSlice[0]].State.Protected = true
			fromPlayer.State.Protected = false // 防bug，不能守自己
			break
		}
	}
	// 判断杀害
	var killed model.Player
	for fromPlayer, toPlayerIndexSlice := range castPoolObj {
		if fromPlayer.Character == Imp && !fromPlayer.State.Poisoned &&
			!game.Players[toPlayerIndexSlice[0]].State.Protected &&
			game.Players[toPlayerIndexSlice[0]].Character != Soldier {
			if game.Players[toPlayerIndexSlice[0]].Character != Mayor {
				// 死的人
				game.Players[toPlayerIndexSlice[0]].State.Dead = true
				killed = game.Players[toPlayerIndexSlice[0]]
			} else {
				// 刀市长
				for {
					randInt := rand.Intn(len(game.Players))
					if !game.Players[randInt].State.Dead &&
						game.Players[randInt].CharacterType != Demons &&
						game.Players[randInt].Character != Mayor {
						// 死的是除了市长和恶魔的其他任意一人
						game.Players[randInt].State.Dead = true
						killed = game.Players[randInt]
						break
					}
				}
			}
			// 自刀
			if game.Players[toPlayerIndexSlice[0]].Character == Imp {
				var scarletWoman *model.Player
				var minionsAlive []*model.Player
				var aliveQuantity int
				for i, player := range game.Players {
					if !player.State.Dead {
						aliveQuantity += 1
					}
					if player.Character == ScarletWoman && !player.State.Dead {
						scarletWoman = &game.Players[i]
					} else {
						if player.CharacterType == Minions && !player.State.Dead {
							minionsAlive = append(minionsAlive, &game.Players[i])
						}
					}
				}
				// 有魅魔且没死
				if !reflect.ValueOf(scarletWoman).IsZero() {
					scarletWoman.CharacterType = Demons
					scarletWoman.Character = Imp
					scarletWoman.State.Evil = true
					scarletWoman.State.Demon = true
					break
				}
				// 如果没有魅魔或魅魔死了
				if reflect.ValueOf(scarletWoman).IsZero() && len(minionsAlive) != 0 {
					randInt := rand.Intn(len(minionsAlive))
					minionsAlive[randInt].CharacterType = Demons
					minionsAlive[randInt].Character = Imp
					minionsAlive[randInt].State.Evil = true
					minionsAlive[randInt].State.Demon = true
					break
				}
			}
			break
		}
	}
	// 结算其他夜晚得知的消息
	if game.State.Stage != 1 && game.State.Stage%2 == 1 {
		for i, player := range game.Players {
			msgPlayer = "您"
			msgAll = ""
			switch player.Character {
			// 给守鸦人提供信息
			case Ravenkeeper:
				if !reflect.ValueOf(killed).IsZero() &&
					!player.State.Drunk && (!player.State.Poisoned || player.State.Protected) &&
					player.Id == killed.Id {
					for fromPlayer, toPlayerIndexSlice := range castPoolObj {
						if fromPlayer.Id == player.Id {
							// 拼接日志
							msgAll += fmt.Sprintf("[%s] ", player.Name)
							info := fmt.Sprintf("发现 [%s] 的身份是 {%s}", game.Players[toPlayerIndexSlice[0]].Name, game.Players[toPlayerIndexSlice[0]].Character)
							msgPlayer += info
							msgAll += info
							game.Players[fromPlayer.Index].Log += msgPlayer + "\n"
							game.Log += msgAll + "\n"
							// 发送日志
							for id, conn := range cfg.ConnPool {
								if id == fromPlayer.Id {
									if err := conn.WriteMessage(websocket.TextMessage, []byte(msgPlayer)); err != nil {
										log.Println("Write error:", err)
										return
									}
									break
								}
							}
							break
						}
					}
				}
			// 给共情者提供信息
			case Empath:
				if !player.State.Dead { // 当晚死亡得不到信息
					evilQuantity := 0 // 记录左右邪恶玩家数量
					if !player.State.Drunk && (!player.State.Poisoned || player.State.Protected) && !player.State.Dead {
						// 生成连座信息
						var left int
						var right int
						if player.Index == 0 {
							left = len(game.Players) - 1
							right = player.Index + 1
						} else if player.Index == len(game.Players)-1 {
							left = player.Index - 1
							right = 0
						} else {
							left = player.Index - 1
							right = player.Index + 1
						}
						for {
							if !game.Players[left].State.Dead && game.Players[left].State.Evil {
								evilQuantity += 1
							}
							if !game.Players[right].State.Dead && game.Players[right].State.Evil {
								evilQuantity += 1
							}
							left--
							if left < 0 {
								left = len(game.Players) - 1
							}
							if left == right {
								break
							}
							right++
							if right > len(game.Players)-1 {
								right = 0
							}
							if left == right {
								break
							}
						}
					} else {
						// 生成伪信息
						randInt := rand.Intn(3)
						evilQuantity = randInt
					}
					// 拼接日志
					msgAll += fmt.Sprintf("[%s] ", player.Name)
					info := fmt.Sprintf("发现与您邻座的邪恶玩家有 {%d} 个", evilQuantity)
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
			// 给掘墓人提供信息
			case Undertaker:
				if !player.State.Dead {
					if !reflect.ValueOf(executed).IsZero() {
						// 无人被处决
						msgAll += fmt.Sprintf("[%s] ", player.Name)
						info := "发现今日无人被处决"
						msgPlayer += info
						msgAll += info
					} else {
						var executedPlayer *model.Player
						if !player.State.Drunk && (!player.State.Poisoned || player.State.Protected) {
							// 生成死亡玩家身份信息
							executedPlayer = executed
						} else {
							// 生成伪信息
							for {
								randInt := rand.Intn(len(game.Players))
								if randInt != player.Index && randInt != executed.Index {
									executedPlayer = &game.Players[randInt]
									break
								}
							}
						}
						// 拼接日志
						msgAll += fmt.Sprintf("[%s] ", player.Name)
						info := fmt.Sprintf("发现今晚被处决的玩家 [%s] 的身份是 {%s}", executed.Name, executedPlayer.Character)
						msgPlayer += info
						msgAll += info
					}
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
	}
	// 判断占卜
	for fromPlayer, toPlayerIndexSlice := range castPoolObj {
		if fromPlayer.Character == FortuneTeller && !fromPlayer.State.Dead {
			msgPlayer = "您"
			msgAll = ""
			var hasDemon = "无"
			if !fromPlayer.State.Drunk && !fromPlayer.State.Poisoned {
				if game.Players[toPlayerIndexSlice[0]].State.Demon || game.Players[toPlayerIndexSlice[1]].State.Demon {
					hasDemon = "有"
				}
			} else {
				randInt := rand.Intn(2)
				if randInt == 0 {
					hasDemon = "无"
				} else {
					hasDemon = "有"
				}
			}
			// 拼接日志
			msgAll += fmt.Sprintf("[%s] ", fromPlayer.Name)
			info := fmt.Sprintf("发现 [%s] 和 [%s] 中 {%s} 恶魔", game.Players[toPlayerIndexSlice[0]].Name, game.Players[toPlayerIndexSlice[1]].Name, hasDemon)
			msgPlayer += info
			msgAll += info
			game.Players[fromPlayer.Index].Log += msgPlayer + "\n"
			game.Log += msgAll + "\n"
			// 发送日志
			for id, conn := range cfg.ConnPool {
				if id == fromPlayer.Id {
					if err := conn.WriteMessage(websocket.TextMessage, []byte(msgPlayer)); err != nil {
						log.Println("Write error:", err)
						return
					}
					break
				}
			}
			break
		}
	}
	// 给管家提供信息
	for fromPlayer, toPlayerIndexSlice := range castPoolObj {
		if fromPlayer.Character == Butler && !fromPlayer.State.Poisoned && !fromPlayer.State.Dead {
			game.Players[toPlayerIndexSlice[0]].State.Vote += 1
			msgPlayer = "您"
			msgAll = ""
			// 拼接日志
			msgAll += fmt.Sprintf("[%s] ", fromPlayer.Name)
			info := fmt.Sprintf("认定 [%s] 为主人，他的投票为两票", game.Players[toPlayerIndexSlice[0]].Name)
			msgPlayer += info
			msgAll += info
			game.Players[fromPlayer.Index].Log += msgPlayer + "\n"
			game.Log += msgAll + "\n"
			// 发送日志
			for id, conn := range cfg.ConnPool {
				if id == fromPlayer.Id {
					if err := conn.WriteMessage(websocket.TextMessage, []byte(msgPlayer)); err != nil {
						log.Println("Write error:", err)
						return
					}
					break
				}
			}
			break
		}
	}
	// 判断本局夜晚结算

}
