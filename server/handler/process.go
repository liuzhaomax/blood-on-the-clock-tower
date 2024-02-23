package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/model"
	"log"
	"math"
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

	mux := &sync.Mutex{}

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway) {
				log.Println("Client disconnected:", err)
				return
			}
			log.Println("Read error:", err)
			conn.Close()
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
		case "nominate":
			nominate(mux, game, playerId, actionReq.Targets)
		case "vote":
			vote(mux, game, playerId, actionReq.Targets)
		case "checkout_night":
			if playerId == game.Host {
				checkoutNight(mux, game, executed)
			}
		case "checkout_day":
			if playerId == game.Host {
				checkoutDay(mux, game)
			}
		case "end_voting":
			if playerId == game.Host {
				executed = endVoting(mux, game)
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
		msg = fmt.Sprintf("第%d天，入夜~\n", game.State.Day)
		// 入夜清除中毒、守护、主人效果
		for i := range game.Players {
			game.Players[i].State.Poisoned = false
			game.Players[i].State.Protected = false
			game.Players[i].State.Master = false
		}
	} else {
		msg = fmt.Sprintf("第%d天，天亮了~\n", game.State.Day)
	}
	// 存入总日志
	game.Log += msg
	// 存入个人日志，刷新的时候加载
	for i := range game.Players {
		game.Players[i].Log += msg
	}
	// 日夜切换
	game.State.Night = !game.State.Night

	for i := range game.Players {
		// 活人调整状态
		if !game.Players[i].State.Dead {
			// 让所有活人重新可以投票，夜转日结算，没投票还有票
			game.Players[i].Ready.Nominated = true
			game.Players[i].Ready.Vote = 1
			game.Players[i].Ready.VoteCount = 0
		}
		// 管家无票
		if game.Players[i].Character == Butler {
			game.Players[i].Ready.Vote = 0
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
	// 第一夜恶魔爪牙互认身份
	if game.State.Stage == 1 {
		msg = ""
		var demon model.Player
		for i, player := range game.Players {
			if player.CharacterType == Demons {
				demon = game.Players[i]
				break
			}
		}
		// 发送爪牙身份给恶魔
		minions := map[string]string{}
		for _, player := range game.Players {
			if player.CharacterType == Minions {
				minions[player.Name] = player.Character
				msg += fmt.Sprintf("本局小恶魔是 [%s]\n", demon.Name)
				if err := cfg.ConnPool[player.Id].WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
					log.Println("Write error:", err)
					return
				}
			}
		}
		// 发送恶魔身份给爪牙
		msg = ""
		for name, character := range minions {
			msg += fmt.Sprintf("本局爪牙 [%s] 的身份是 [%s]\n", name, character)
		}
		if err := cfg.ConnPool[demon.Id].WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			log.Println("Write error:", err)
			return
		}
	}
}

func endVoting(mux *sync.Mutex, game *model.Room) (executed *model.Player) {
	cfg := model.GetConfig()
	mux.Lock()
	defer mux.Unlock()

	var msg string

	var expectantExecuted *model.Player // 被投票者（被提名者）
	var aliveCount int                  // 活人数量
	for i, player := range game.Players {
		if player.Ready.VoteCount > 0 {
			expectantExecuted = &game.Players[i]
		}
		if !player.State.Dead {
			aliveCount++
		}
	}
	if expectantExecuted != nil && expectantExecuted.Ready.VoteCount > int(math.Floor(float64(aliveCount/2))) {
		executed = expectantExecuted
		executed.State.Dead = true
		executed.Ready.Vote = 1 // 死人还有一票
		executed.Ready.Nominate = false
		executed.Ready.Nominated = false
		executed.Ready.VoteCount = 0
		msg += fmt.Sprintf("处决结果：[%s] 被公投处决，死亡\n", executed.Name)
	} else {
		msg += "处决结果：无人被处决\n"
	}
	// 发送日志
	for _, conn := range cfg.ConnPool {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			log.Println("Write error:", err)
			return
		}
	}
	// 判断圣徒 邪恶胜利条件4
	if executed != nil && executed.Character == Saint {
		checkout(game, executed)
	}
	// 退出投票环节
	game.State.VotingStep = false
	return
}

func nominate(mux *sync.Mutex, game *model.Room, playerId string, targets []string) {
	cfg := model.GetConfig()
	mux.Lock()
	defer mux.Unlock()

	var msg = ""
	var msgName = ""

	for i, player := range game.Players {
		if player.Id == playerId && player.Ready.Nominate && !player.State.Dead {
			msg += fmt.Sprintf("[%s] ", player.Name)
			msgName = msg
			for j, player := range game.Players {
				if targets[0] == player.Id && player.Ready.Nominated && !player.State.Dead {
					game.Players[i].Ready.Nominate = false
					game.Players[j].Ready.Nominated = false
					msg += fmt.Sprintf("提名 [%s] 进行处决公投\n", player.Name)
					break
				}
			}
			break
		}
	}
	for i := range game.Players {
		game.Players[i].Log += msg
	}
	game.Log += msg
	// 发送日志
	for _, conn := range cfg.ConnPool {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			log.Println("Write error:", err)
			return
		}
	}

	// 判断圣女
	var canGoToVotingStep = true
	msg = msgName
	for i, player := range game.Players {
		if player.Character == Virgin && player.Id == targets[0] && player.State.Blessed {
			game.Players[i].State.Blessed = false
			for i, player := range game.Players {
				if player.Id == playerId {
					game.Players[i].State.Dead = true
					break
				}
			}
			canGoToVotingStep = false
			msg += "被圣女弹死了\n"
			break
		}
	}
	for i := range game.Players {
		game.Players[i].Log += msg
	}
	game.Log += msg
	// 发送日志
	for _, conn := range cfg.ConnPool {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			log.Println("Write error:", err)
			return
		}
	}

	// 进入投票环节 圣女条件成立就不用进投票环节
	if canGoToVotingStep {
		game.State.VotingStep = true
	}
}

func vote(mux *sync.Mutex, game *model.Room, playerId string, targets []string) {
	cfg := model.GetConfig()
	mux.Lock()
	defer mux.Unlock()

	var msg = ""

	for i, player := range game.Players {
		if player.Id == playerId && player.Ready.Vote > 0 {
			var masterFlag bool // 是否是管家的主人
			if game.Players[i].Ready.Vote == 2 {
				masterFlag = true
			}
			msg += fmt.Sprintf("[%s] ", player.Name)
			for k, player := range game.Players {
				if targets[0] == player.Id && !player.State.Dead {
					game.Players[i].Ready.Vote = 0
					game.Players[k].Ready.VoteCount += 1
					msg += fmt.Sprintf("决意投票给 [%s] \n", player.Name)
					if masterFlag {
						game.Players[k].Ready.VoteCount += 1 // 是管家的主人再加1票
						for j := range game.Players {
							if game.Players[j].Character == Butler {
								msg += fmt.Sprintf("\n[%s] 决意投票给 [%s] \n", game.Players[j].Name, player.Name)
								break
							}
						}
					}
					break
				}
			}
			break
		}
	}
	for i := range game.Players {
		game.Players[i].Log += msg
	}
	game.Log += msg
	// 发送日志
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
		if player.Id == playerId && !player.State.Dead {
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

func checkoutNight(mux *sync.Mutex, game *model.Room, executed *model.Player) {
	cfg := model.GetConfig()
	mux.Lock()
	defer mux.Unlock()

	var msgPlayer = "您"
	var msgAll = ""

	// 承载技能释放者对象的池
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

	// 各身份夜晚技能结算-有顺序
	// 判断下毒
	for fromPlayer, toPlayerIndexSlice := range castPoolObj {
		if fromPlayer.Character == Poisoner && !fromPlayer.State.Dead {
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
						if game.Players[randInt].CharacterType == Townsfolk && game.Players[randInt].Id != player.Id {
							realFolk = game.Players[randInt]
							break
						}
					}
					for {
						randInt := rand.Intn(len(game.Players))
						if game.Players[randInt].Id != realFolk.Id && game.Players[randInt].Id != player.Id {
							other = game.Players[randInt]
							break
						}
					}
				} else {
					// 生成伪信息
					for {
						randInt := rand.Intn(len(game.Players))
						if game.Players[randInt].Id != player.Id {
							realFolk = game.Players[randInt]
							break
						}
					}
					for {
						randInt := rand.Intn(len(game.Players))
						if game.Players[randInt].Id != realFolk.Id && game.Players[randInt].Id != player.Id {
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
					info = fmt.Sprintf("发现 [%s] 和 [%s] 其中一个是 {%s}\n", realFolk.Name, other.Name, realFolk.Character)
				} else {
					info = fmt.Sprintf("发现 [%s] 和 [%s] 其中一个是 {%s}\n", other.Name, realFolk.Name, realFolk.Character)
				}
				msgPlayer += info
				msgAll += info
				game.Players[i].Log += msgPlayer
				game.Log += msgAll
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
							if game.Players[randInt].CharacterType == Outsiders && game.Players[randInt].Id != player.Id {
								realOutsider = game.Players[randInt]
								break
							}
						}
						for {
							randInt := rand.Intn(len(game.Players))
							if game.Players[randInt].Id != realOutsider.Id && game.Players[randInt].Id != player.Id {
								other = game.Players[randInt]
								break
							}
						}
					}
				} else {
					// 生成伪信息
					var hasOutsider bool
					randInt := rand.Intn(len(game.Players))
					if game.Players[randInt].Id != player.Id {
						realOutsider = game.Players[randInt]
						hasOutsider = true
					}
					if hasOutsider {
						for {
							randInt := rand.Intn(len(game.Players))
							if game.Players[randInt].Id != realOutsider.Id && game.Players[randInt].Id != player.Id {
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
						info = fmt.Sprintf("发现 [%s] 和 [%s] 其中一个是 {%s}\n", realOutsider.Name, other.Name, realOutsider.Character)
					} else {
						info = fmt.Sprintf("发现 [%s] 和 [%s] 其中一个是 {%s}\n", other.Name, realOutsider.Name, realOutsider.Character)
					}
				}
				msgPlayer += info
				msgAll += info
				game.Players[i].Log += msgPlayer
				game.Log += msgAll
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
						if game.Players[randInt].CharacterType == Minions && game.Players[randInt].Id != player.Id {
							realMinion = game.Players[randInt]
							break
						}
					}
					for {
						randInt := rand.Intn(len(game.Players))
						if game.Players[randInt].Id != realMinion.Id && game.Players[randInt].Id != player.Id {
							other = game.Players[randInt]
							break
						}
					}
				} else {
					// 生成伪信息
					for {
						randInt := rand.Intn(len(game.Players))
						if game.Players[randInt].Id != player.Id {
							realMinion = game.Players[randInt]
							break
						}
					}
					for {
						randInt := rand.Intn(len(game.Players))
						if game.Players[randInt].Id != realMinion.Id && game.Players[randInt].Id != player.Id {
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
					info = fmt.Sprintf("发现 [%s] 和 [%s] 其中一个是 {%s}\n", realMinion.Name, other.Name, realMinion.Character)
				} else {
					info = fmt.Sprintf("发现 [%s] 和 [%s] 其中一个是 {%s}\n", other.Name, realMinion.Name, realMinion.Character)
				}
				msgPlayer += info
				msgAll += info
				game.Players[i].Log += msgPlayer
				game.Log += msgAll
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
				info := fmt.Sprintf("发现互为邻座的邪恶玩家有 {%d} 对\n", connected)
				msgPlayer += info
				msgAll += info
				game.Players[i].Log += msgPlayer
				game.Log += msgAll
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
				info := fmt.Sprintf("发现与您邻座的邪恶玩家有 {%d} 个\n", evilQuantity)
				msgPlayer += info
				msgAll += info
				game.Players[i].Log += msgPlayer
				game.Log += msgAll
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
				msgAll += fmt.Sprintf("[%s] 是间谍，知晓所有身份\n", player.Name)
				var info string
				for _, player := range game.Players {
					info += fmt.Sprintf("知晓所有身份：\n玩家 [%s] 的身份是 {%s}\n", player.Name, player.Character)
				}
				msgPlayer += info
				game.Players[i].Log += msgPlayer
				game.Log += msgAll
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
	var killed *model.Player
	for fromPlayer, toPlayerIndexSlice := range castPoolObj {
		if fromPlayer.Character == Imp && !fromPlayer.State.Poisoned &&
			!game.Players[toPlayerIndexSlice[0]].State.Protected &&
			game.Players[toPlayerIndexSlice[0]].Character != Soldier {
			if game.Players[toPlayerIndexSlice[0]].Character != Mayor {
				// 死的人
				game.Players[toPlayerIndexSlice[0]].State.Dead = true
				killed = &game.Players[toPlayerIndexSlice[0]]
			} else {
				// 刀市长
				for {
					randInt := rand.Intn(len(game.Players))
					if !game.Players[randInt].State.Dead &&
						game.Players[randInt].CharacterType != Demons {
						// 死的是除了恶魔的其他任意一人
						game.Players[randInt].State.Dead = true
						killed = &game.Players[randInt]
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
				if killed != nil &&
					!player.State.Drunk && (!player.State.Poisoned || player.State.Protected) &&
					player.Id == killed.Id {
					for fromPlayer, toPlayerIndexSlice := range castPoolObj {
						if fromPlayer.Id == player.Id {
							// 拼接日志
							msgAll += fmt.Sprintf("[%s] ", player.Name)
							info := fmt.Sprintf("发现 [%s] 的身份是 {%s}\n", game.Players[toPlayerIndexSlice[0]].Name, game.Players[toPlayerIndexSlice[0]].Character)
							msgPlayer += info
							msgAll += info
							game.Players[fromPlayer.Index].Log += msgPlayer
							game.Log += msgAll
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
					info := fmt.Sprintf("发现与您邻座的邪恶玩家有 {%d} 个\n", evilQuantity)
					msgPlayer += info
					msgAll += info
					game.Players[i].Log += msgPlayer
					game.Log += msgAll
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
					if executed == nil {
						// 无人被处决
						msgAll += fmt.Sprintf("[%s] ", player.Name)
						info := "发现今日无人被处决\n"
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
						info := fmt.Sprintf("发现今晚被处决的玩家 [%s] 的身份是 {%s}\n", executed.Name, executedPlayer.Character)
						msgPlayer += info
						msgAll += info
					}
					game.Players[i].Log += msgPlayer
					game.Log += msgAll
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
			info := fmt.Sprintf("发现 [%s] 和 [%s] 中 {%s} 恶魔\n", game.Players[toPlayerIndexSlice[0]].Name, game.Players[toPlayerIndexSlice[1]].Name, hasDemon)
			msgPlayer += info
			msgAll += info
			game.Players[fromPlayer.Index].Log += msgPlayer
			game.Log += msgAll
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
			game.Players[toPlayerIndexSlice[0]].State.Master = true
			game.Players[toPlayerIndexSlice[0]].Ready.Vote += 1
			msgPlayer = "您"
			msgAll = ""
			// 拼接日志
			msgAll += fmt.Sprintf("[%s] ", fromPlayer.Name)
			info := fmt.Sprintf("认定 [%s] 为主人，他本轮拥有两票\n", game.Players[toPlayerIndexSlice[0]].Name)
			msgPlayer += info
			msgAll += info
			game.Players[fromPlayer.Index].Log += msgPlayer
			game.Log += msgAll
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

	// ------- 结算 -------
	msg := ""
	// 结算本阶段
	if killed == nil {
		msg += "昨夜是 平安夜\n"
	} else {
		msg += fmt.Sprintf("昨夜 [%s] 死亡\n", killed.Name)
	}
	// 发送日志
	for _, conn := range cfg.ConnPool {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			log.Println("Write error:", err)
			return
		}
	}
	// 结算本局
	checkout(game, nil) // 这里不要传入executed，因为晚上不处决人，晚上可能死圣徒
}

func checkoutDay(mux *sync.Mutex, game *model.Room) {
	cfg := model.GetConfig()
	mux.Lock()
	defer mux.Unlock()

	var msgPlayer = "您"
	var msgAll = ""

	// 承载技能释放者对象的池
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

	// 判断杀手
	for fromPlayer, toPlayerIndexSlice := range castPoolObj {
		if fromPlayer.Character == Slayer && fromPlayer.State.Bullet && len(toPlayerIndexSlice) == 1 {
			info := ""
			msgAll += fmt.Sprintf("[%s] ", fromPlayer.Name)
			game.Players[fromPlayer.Index].State.Bullet = false
			if game.Players[toPlayerIndexSlice[0]].CharacterType == Demons {
				game.Players[toPlayerIndexSlice[0]].State.Dead = true
				// 拼接日志
				info += fmt.Sprintf("发现 [%s] 被您枪杀了\n", game.Players[toPlayerIndexSlice[0]].Name)
			} else {
				// 拼接日志
				info += "发现没人被您枪杀\n"
			}
			msgPlayer += info
			msgAll += info
			game.Players[fromPlayer.Index].Log += msgPlayer
			game.Log += msgAll
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

	// 结算本局
	checkout(game, nil)
}

// checkout 结算本局
func checkout(game *model.Room, executed *model.Player) {
	cfg := model.GetConfig()
	msg := ""
	var realDemonCount int   // 恶魔数量，被占卜认定的不算
	var canNominate int      // 可提名人数
	var hasSlayerBullet bool // 有枪手且枪手有子弹
	var aliveCount int       // 活人数量
	var canVote int          // 可投票数量
	var evilCount int        // 邪恶玩家数量
	for _, player := range game.Players {
		// 对应邪恶胜利条件1
		if player.Character == Slayer && player.State.Bullet {
			hasSlayerBullet = true
		}
		if player.Ready.Nominate {
			canNominate++
		}
		// 对应邪恶胜利条件2
		if player.Ready.Vote > 0 {
			canVote += player.Ready.Vote
		}
		if player.State.Evil && player.Character != Recluse {
			evilCount++
		}
		// 对应邪恶胜利条件3
		if !player.State.Dead {
			aliveCount++
		}
		// 对应平民胜利条件
		if player.CharacterType == Demons {
			realDemonCount++
		}
	}
	// 邪恶胜利条件1
	if !hasSlayerBullet && canNominate == 0 {
		msg += "本局结束，邪恶胜利\n"
		game.Result = "邪恶阵营胜利"
	}
	// 邪恶胜利条件2
	halfAlive := int(math.Floor(float64(aliveCount / 2)))
	if canVote <= halfAlive && evilCount >= halfAlive {
		msg += "本局结束，邪恶胜利\n"
		game.Result = "邪恶阵营胜利"
	}
	// 邪恶胜利条件3
	if evilCount == aliveCount {
		msg += "本局结束，邪恶胜利\n"
		game.Result = "邪恶阵营胜利"
	}
	// 邪恶胜利条件4
	if executed != nil && executed.Character == Saint {
		msg += "本局结束，邪恶胜利\n"
		game.Result = "邪恶阵营胜利"
	}
	// 平民胜利条件（恶魔受不了了自杀情况）
	if realDemonCount == 0 {
		msg += "本局结束，平民胜利\n"
		game.Result = "平民阵营胜利"
	}
	// 拼接日志
	for i := range game.Players {
		game.Players[i].Log += msg
	}
	game.Log += msg
	// 发送日志
	for _, conn := range cfg.ConnPool {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			log.Println("Write error:", err)
			return
		}
	}
}
