package handler

import (
	"fmt"
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/model"
	"sync"
)

func cast(mux *sync.Mutex, game *model.Room, playerId string, targets []string) {
	mux.Lock()
	defer mux.Unlock()

	var msgPlayer = "您"
	var msgAll = ""
	var slayerTarget *model.Player
	for i, player := range game.Players {
		if player.Id == playerId && !player.State.Dead {
			msgAll += fmt.Sprintf("[%s] ", player.Name)
			switch player.Character {
			case Poisoner:
				for _, player := range game.Players {
					if targets[0] == player.Id {
						info := fmt.Sprintf("对 [%s] 进行了投毒！\n", player.Name)
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
						info := player.Name + "] 进行了卜算！\n"
						msgPlayer += info
						msgAll += info
						break
					}
				}
			case Butler:
				for _, player := range game.Players {
					if targets[0] == player.Id {
						info := fmt.Sprintf(" 对 [%s] 进行了认主！\n", player.Name)
						msgPlayer += info
						msgAll += info
						break
					}
				}
			case Monk:
				for _, player := range game.Players {
					if targets[0] == player.Id {
						info := fmt.Sprintf(" 对 [%s] 进行了守护！\n", player.Name)
						msgPlayer += info
						msgAll += info
						break
					}
				}
			case Imp:
				for _, player := range game.Players {
					if targets[0] == player.Id {
						info := fmt.Sprintf(" 对 [%s] 进行了杀害！\n", player.Name)
						msgPlayer += info
						msgAll += info
						break
					}
				}
			case Slayer:
				for i, player := range game.Players {
					if targets[0] == player.Id {
						slayerTarget = &game.Players[i]
						info := fmt.Sprintf(" 对 [%s] 进行了枪毙！\n", player.Name)
						msgPlayer += info
						msgAll += info
						break
					}
				}
			case Ravenkeeper:
				for _, player := range game.Players {
					if targets[0] == player.Id {
						info := fmt.Sprintf(" 对 [%s] 进行了反向通灵，如果施法者没死，则技能无效！\n", player.Name)
						msgPlayer += info
						msgAll += info
						break
					}
				}
			}
			game.Players[i].State.Casted = true
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
	game.Log += msgAll
	// 发送日志
	emit(game, playerId)
	// 发送game 更新casted状态
	broadcast(game)

	// 判断杀手
	if slayerTarget != nil {
		msgPlayer = "您"
		msgAll = ""
		for i, player := range game.Players {
			if player.Character == Slayer {
				msgAll += fmt.Sprintf("[%s] ", player.Name)
				// 考虑子弹，有子弹才有后续判定日志，没子弹不会有成功失败提示
				if game.Players[i].State.Bullet {
					game.Players[i].State.Bullet = false // 子弹不管怎样都会发射
					if (slayerTarget.CharacterType == Demons || slayerTarget.State.RegardedAs == Imp) &&
						!player.State.Drunk && !player.State.Poisoned {
						slayerTarget.State.Dead = true
						slayerTarget.Ready.Nominate = false
						slayerTarget.Ready.Nominated = false
						// 拼接日志
						msgAll += fmt.Sprintf("枪杀了 [%s] \n", slayerTarget.Name)
						for i := range game.Players {
							game.Players[i].Log += msgAll
						}
						game.Log += msgAll
						// 发送日志
						broadcast(game)
						// 判断魅魔 - 小恶魔被枪毙，活人大于等于5个，有魅魔且没死
						var aliveCount int // 活人数量
						var scarletWomanIndex = -1
						for _, player := range game.Players {
							if !player.State.Dead {
								aliveCount++
							}
							if player.Character == ScarletWoman && !player.State.Dead {
								scarletWomanIndex = player.Index
							}
						}
						if slayerTarget.Character == Imp && aliveCount >= 5 &&
							scarletWomanIndex > 0 && !game.Players[scarletWomanIndex].State.Poisoned {
							scarletWoman := &game.Players[scarletWomanIndex]
							scarletWoman.Character = Imp
							scarletWoman.CharacterType = Demons
							scarletWoman.State.Demon = true
							// 拼接日志
							msgPlayer := "您"
							msgAll := ""
							info := "已变为 {小恶魔} \n"
							msgPlayer += info
							msgAll += fmt.Sprintf("[%s] ", scarletWoman.Name) + info
							scarletWoman.Log += msgPlayer
							game.Log += msgAll
							// 发送日志
							emit(game, scarletWoman.Id)
						}
						// 立即结算
						checkout(game, nil)
						break
					} else {
						// 拼接日志
						info := "枪杀失败，无事发生\n"
						msgPlayer += info
						msgAll += info
						game.Players[i].Log += msgPlayer
						game.Log += msgAll
						// 发送日志
						emit(game, player.Id)
					}
					break
				} else {
					// 拼接日志
					info := "空枪，无事发生\n"
					msgPlayer += info
					msgAll += info
					game.Players[i].Log += msgPlayer
					game.Log += msgAll
					// 发送日志
					emit(game, player.Id)
					break
				}
			}
		}
	}
}
