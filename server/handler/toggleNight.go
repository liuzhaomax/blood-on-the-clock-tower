package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/model"
	"log"
	"math/rand"
	"sync"
)

func toggleNight(mux *sync.Mutex, game *model.Room) {
	mux.Lock()
	defer mux.Unlock()

	var msg string

	if game.Result == "" {
		// Stage + 1
		game.State.Stage += 1
		// 夜转日 Day+1
		if !game.State.Night {
			msg = fmt.Sprintf("第%d天，入夜~\n", game.State.Day+1)
			// 入夜清除中毒、守护、主人效果
			for i := range game.Players {
				game.Players[i].State.Poisoned = false
				game.Players[i].State.Protected = false
				game.Players[i].State.Master = false
				// 还原隐士邪恶身份 - 清除中毒效果
				if game.Players[i].Character == Recluse {
					game.Players[i].State.Evil = true
					if game.Players[i].State.RegardedAsSaved == Spy {
						game.Players[i].State.Evil = false
					}
					game.Players[i].State.RegardedAs = game.Players[i].State.RegardedAsSaved
					if game.Players[i].State.RegardedAsSaved == Imp {
						game.Players[i].State.Demon = true
					}
				}
				// 还原间谍平民身份 - 清除中毒效果
				if game.Players[i].Character == Spy {
					game.Players[i].State.Evil = false
				}
			}
		} else {
			game.State.Day += 1
			msg = fmt.Sprintf("第%d天，天亮了~\n", game.State.Day+1)
			// 天亮 清空executed
			game.Executed = nil
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
			// 活人调整状态 - 让所有活人重新可以投票，夜转日结算，没投票还有票
			if !game.Players[i].State.Dead {
				game.Players[i].Ready.Nominate = true
				game.Players[i].Ready.Nominated = true
			}
			// 调整玩家施放技能的准备状态
			game.Players[i].State.Casted = true
			game.CastPool = map[string][]string{}
			if !game.Players[i].State.Dead {
				switch game.Players[i].Character {
				case Poisoner:
					if game.State.Stage%2 == 1 && !game.Players[i].State.Dead {
						game.Players[i].State.Casted = false
					}
				case FortuneTeller:
					if game.State.Stage%2 == 1 && !game.Players[i].State.Dead {
						game.Players[i].State.Casted = false
					}
				case Butler:
					if game.State.Stage%2 == 1 && !game.Players[i].State.Dead {
						game.Players[i].State.Casted = false
					}
				case Monk:
					if game.State.Stage%2 == 1 && game.State.Stage != 1 && !game.Players[i].State.Dead {
						game.Players[i].State.Casted = false
					}
				case Imp:
					if game.State.Stage%2 == 1 && game.State.Stage != 1 && !game.Players[i].State.Dead {
						game.Players[i].State.Casted = false
					}
				case Ravenkeeper:
					if game.State.Stage%2 == 1 && game.State.Stage != 1 && !game.Players[i].State.Dead {
						game.Players[i].State.Casted = false
					}
				case Slayer:
					if game.State.Stage%2 == 0 && !game.Players[i].State.Dead && game.Players[i].State.Bullet {
						game.Players[i].State.Casted = false
					}
				default:
					game.Players[i].State.Casted = true
				}
			}
			// 让所有人的僧侣守护状态消失
			game.Players[i].State.Protected = false
		}
		// 将日夜切换日志群发
		broadcast(game)
		// 为间谍重新分配注册身份类型
		if game.State.Night {
			for i, player := range game.Players {
				if player.Character == Spy && !player.State.Dead {
					msgAll := fmt.Sprintf("[%s] ", player.Name)
					msgPlayer := "您"
					info := ""
					registeredAsTypePool := []string{Townsfolk, Outsiders}
					randIdx := rand.Intn(len(registeredAsTypePool))
					game.Players[i].State.RegisteredAsType = registeredAsTypePool[randIdx]
					randIdxChar := -1
					if registeredAsTypePool[randIdx] == Townsfolk {
						randIdxChar = rand.Intn(len(TownsfolkPool))
						game.Players[i].State.RegardedAs = TownsfolkPool[randIdxChar]
					}
					if registeredAsTypePool[randIdx] == Outsiders {
						randIdxChar = rand.Intn(len(OutsidersPool))
						game.Players[i].State.RegardedAs = OutsidersPool[randIdxChar]
					}
					// 拼接日志
					info += fmt.Sprintf("今日注册的身份是 {%s} \n", game.Players[i].State.RegardedAs)
					msgPlayer += info
					msgAll += info
					game.Players[i].Log += msgPlayer
					game.Log += msgAll
					// 发送
					emit(game, player.Id)
					break
				}
			}
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
		for i, player := range game.Players {
			if player.CharacterType == Minions {
				minions[player.Name] = player.Character
				msg += fmt.Sprintf("您发现恶魔 [%s] 的身份是 {%s}\n", demon.Name, demon.Character)
				game.Players[i].Log += msg
				// 发送日志
				emit(game, player.Id)
				break
			}
		}
		// 发送恶魔身份给爪牙
		msg = ""
		for name := range minions {
			msg += fmt.Sprintf("您发现 [%s] 是爪牙\n", name)
		}
		for i, player := range game.Players {
			if player.CharacterType == Demons {
				game.Players[i].Log += msg
				break
			}
		}
		// 保存到总日志
		game.Log += "恶魔爪牙已互认身份\n"
		// 发送
		emit(game, demon.Id)

		// 给恶魔提供3个不在场的村民身份
		msg = findThreeCharactersNotInGame(game.Players)
		game.Players[demon.Index].Log += msg
		// 保存到总日志
		game.Log += msg
		// 发送
		emit(game, demon.Id)
	}
}

func findThreeCharactersNotInGame(players []model.Player) string {
	if len(players) < 7 {
		return "人数小于七人，无法提供三个不在场的村民身份\n"
	}
	hasRepeatedCharacter := false
	round := 0
	var chars []string
	msg := "您发现这三个村民身份不在本局中："
	for {
		hasRepeatedCharacter = false
		randInt := rand.Intn(len(TownsfolkPool))
		for _, player := range players {
			// 和所有场上村民比，是否有重复
			if player.Character == TownsfolkPool[randInt] {
				hasRepeatedCharacter = true
			}
			// 和三个身份比，是否有重复
			for _, char := range chars {
				if char == TownsfolkPool[randInt] {
					hasRepeatedCharacter = true
					break
				}
			}
			// 有重复，则跳出重新随机一个身份
			if hasRepeatedCharacter {
				break
			}
		}
		if !hasRepeatedCharacter {
			round += 1
			chars = append(chars, TownsfolkPool[randInt])
		}
		if round == 3 {
			break
		}
	}
	for _, character := range chars {
		msg += fmt.Sprintf("{%s} ", character)
	}
	return msg + "\n"
}

// broadcast 广播game
func broadcast(game *model.Room) {
	game.ResMux.Lock()
	defer game.ResMux.Unlock()
	marshaledGame, err := json.Marshal(*game)
	if err != nil {
		log.Println("JSON marshal error:", err)
		return
	}
	game.GameConnPool.Range(func(id, conn any) bool {
		if err = conn.(*websocket.Conn).WriteMessage(websocket.TextMessage, marshaledGame); err != nil {
			log.Println("Write error:", err)
			return false
		}
		return true
	})
}

// emit 发送game到指定终端
func emit(game *model.Room, destinationId string) {
	game.ResMux.Lock()
	defer game.ResMux.Unlock()
	marshaledGame, err := json.Marshal(*game)
	if err != nil {
		log.Println("JSON marshal error:", err)
		return
	}
	game.GameConnPool.Range(func(id, conn any) bool {
		if id == destinationId {
			if err := conn.(*websocket.Conn).WriteMessage(websocket.TextMessage, marshaledGame); err != nil {
				log.Println("Write error:", err)
				return false
			}
			return false
		}
		return true
	})
}
