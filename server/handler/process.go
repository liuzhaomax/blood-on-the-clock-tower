package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/model"
	"log"
	"math"
	"math/rand"
	"reflect"
	"sync"
)

func toggleNight(mux *sync.Mutex, game *model.Room) {
	mux.Lock()
	defer mux.Unlock()

	var msg string

	if game.Result == "" {
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
			msg = fmt.Sprintf("第%d天，天亮了~\n", game.State.Day)
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
				game.Players[i].Ready.Vote = 1
			}
			game.Players[i].State.Nominated = false
			game.Players[i].State.Voted = false
			game.Players[i].State.VoteCount = 0
			// 调整玩家施放技能的准备状态
			game.Players[i].State.Casted = true
			game.CastPool = map[string][]string{}
			if !game.Players[i].State.Dead {
				switch game.Players[i].Character {
				case Poisoner:
					if game.State.Stage%2 == 1 {
						game.Players[i].State.Casted = false
					}
				case FortuneTeller:
					if game.State.Stage%2 == 1 {
						game.Players[i].State.Casted = false
					}
				case Butler:
					if game.State.Stage%2 == 1 {
						game.Players[i].State.Casted = false
					}
				case Monk:
					if game.State.Stage%2 == 1 && game.State.Stage != 1 {
						game.Players[i].State.Casted = false
					}
				case Imp:
					if game.State.Stage%2 == 1 && game.State.Stage != 1 {
						game.Players[i].State.Casted = false
					}
				case Slayer:
					if game.State.Stage%2 == 0 && game.Players[i].State.Bullet {
						game.Players[i].State.Casted = false
					}
				case Ravenkeeper:
					if game.State.Stage%2 == 1 && game.State.Stage != 1 {
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
		for name, character := range minions {
			msg += fmt.Sprintf("您发现 [%s] 是爪牙\n", name)
			if character == Spy {
				msg += fmt.Sprintf("其身份是 {%s}\n", character)
			}
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

func endVoting(mux *sync.Mutex, game *model.Room) {
	mux.Lock()
	defer mux.Unlock()

	if !game.State.VotingStep {
		return
	}

	// 判断管家 - 看管家的票是否算数
	var hasMasterVoted bool        // 主人是否投了票
	var hasButlerVoted bool        // 管家是否投了票
	var butlerPlayer *model.Player // 管家玩家
	for i, player := range game.Players {
		if player.State.Master && player.State.Voted {
			hasMasterVoted = true
		}
		if player.Character == Butler && player.State.Voted {
			hasButlerVoted = true
			butlerPlayer = &game.Players[i]
		}
	}

	var msg string

	var nominated *model.Player // 被投票者（被提名者）
	var aliveCount int          // 活人数量
	var scarletWomanIndex = -1
	for i, player := range game.Players {
		if player.State.Nominated {
			nominated = &game.Players[i]
		}
		if !player.State.Dead {
			aliveCount++
		}
		if player.Character == ScarletWoman && !player.State.Dead {
			scarletWomanIndex = player.Index
		}
	}
	// 主人没投，管家投了，票数减一，因为如果主人不投票，则管家不能跟票
	if !hasMasterVoted && hasButlerVoted {
		nominated.State.VoteCount -= 1
		butlerPlayer.State.Voted = false
		butlerPlayer.Ready.Vote = 1
		// 拼接日志
		msgButler := fmt.Sprintf("主人未投票，您投给 [%s] 的票无效！\n", nominated.Name)
		butlerPlayer.Log += msgButler
		// 发送日志 - 告诉管家投票无效
		emit(game, butlerPlayer.Id)
		// 票池要清除管家
		for playerId := range game.VotePool {
			if playerId == butlerPlayer.Id {
				game.VotePool[playerId] = ""
				break
			}
		}
	}
	// 主人投了，管家也投了，要加入管家投票成功日志，对应vote函数的管家逻辑
	if hasMasterVoted && hasButlerVoted {
		msg += fmt.Sprintf("[%s] 投票 [%s] 成功\n", butlerPlayer.Name, nominated.Name)
	}
	// 打印所有投票成功的票型
	for _, info := range game.VotePool {
		msg += info
	}
	if nominated != nil && nominated.State.VoteCount > int(math.Floor(float64(aliveCount/2))) {
		game.Executed = nominated
		game.Executed.State.Dead = true
		game.Executed.Ready.Vote = 1 // 死人还有一票
		game.Executed.Ready.Nominate = false
		game.Executed.Ready.Nominated = false
		game.Executed.State.VoteCount = 0
		msg += fmt.Sprintf("处决结果：[%s] 被公投处决，死亡\n", game.Executed.Name)
	} else {
		if nominated != nil {
			nominated.State.Nominated = false
		}
		msg += "处决结果：无人被处决\n"
	}
	for i := range game.Players {
		game.Players[i].Log += msg
	}
	game.Log += msg
	// 发送日志
	broadcast(game)
	// 判断圣徒 - 邪恶胜利条件4
	if game.Executed != nil && game.Executed.Character == Saint && !game.Executed.State.Poisoned {
		checkout(game, game.Executed)
	}
	// 判断魅魔 - 有人被处决，处决的人是小恶魔，活人大于等于5个，有魅魔且没死
	if game.Executed != nil && game.Executed.Character == Imp && aliveCount-1 >= 5 &&
		scarletWomanIndex > 0 && !game.Players[scarletWomanIndex].State.Poisoned {
		scarletWoman := &game.Players[scarletWomanIndex]
		// 拼接日志
		msgPlayer := "您"
		msgAll := ""
		info := "已变为小恶魔\n"
		msgPlayer += info
		msgAll += fmt.Sprintf("[%s] ", scarletWoman.Name) + info
		scarletWoman.Log += msgPlayer
		game.Log += msgAll
		// 发送日志
		emit(game, scarletWoman.Id)
	}
	// 退出投票环节，清空票池
	game.State.VotingStep = false
	game.VotePool = map[string]string{}
	// 发送votingStep
	broadcast(game)
	// 立即结算
	checkout(game, game.Executed)
}

func nominate(mux *sync.Mutex, game *model.Room, playerId string, targets []string) {
	mux.Lock()
	defer mux.Unlock()

	// 如果有处决者产生 不能提名
	if game.Executed != nil {
		msgPlayer := "本轮已处决过人，您的提名无效\n"
		for i, player := range game.Players {
			if player.Id == playerId {
				game.Players[i].Log += msgPlayer
				break
			}
		}
		// 发送日志
		emit(game, playerId)
		return
	}

	var msg = ""
	var msgName = ""

	for i, player := range game.Players {
		if player.Id == playerId && player.Ready.Nominate && !player.State.Dead && !game.State.VotingStep {
			msg += fmt.Sprintf("[%s] ", player.Name)
			msgName = msg
			for j, player := range game.Players {
				if targets[0] == player.Id && player.Ready.Nominated && !player.State.Dead {
					game.Players[i].Ready.Nominate = false  // 发动提名者不能再提名
					game.Players[j].Ready.Nominated = false // 被提名者不能再被提名
					game.Players[j].State.Nominated = true  // 被提名者已被提名
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
	broadcast(game)

	// 判断圣女
	var canGoToVotingStep = true
	for i, player := range game.Players {
		if player.Character == Virgin && player.Id == targets[0] && player.State.Blessed && !player.State.Drunk {
			game.Players[i].State.Blessed = false
			if !player.State.Poisoned {
				msg = msgName
				for i, player := range game.Players {
					if player.Id == playerId && player.CharacterType == Townsfolk && !player.State.Drunk {
						game.Players[i].State.Dead = true
						game.Players[i].Ready.Nominate = false
						game.Players[i].Ready.Nominated = false
						canGoToVotingStep = false
						msg += "被圣女反弹死了\n"
						for i := range game.Players {
							game.Players[i].Log += msg
						}
						game.Log += msg
						// 发送日志
						broadcast(game)
						// 反弹死了算是已有处决人
						game.Executed = &game.Players[i]
						break
					}
				}
				break
			}
		}
	}

	// 进入投票环节 圣女条件成立就不用进投票环节
	if canGoToVotingStep {
		game.State.VotingStep = true
	}
	// 广播votingStep
	broadcast(game)
}

func vote(mux *sync.Mutex, game *model.Room, playerId string) {
	mux.Lock()
	defer mux.Unlock()

	var msgAll = ""
	var msgPlayer = "您"

	for i, player := range game.Players {
		if player.Id == playerId && player.Ready.Vote > 0 && game.State.VotingStep {
			msgAll += fmt.Sprintf("[%s] ", player.Name)
			var nominated *model.Player
			for j, player := range game.Players {
				if player.State.Nominated {
					nominated = &game.Players[j]
					break
				}
			}
			if nominated != nil && !nominated.State.Dead {
				game.Players[i].Ready.Vote = 0
				game.Players[i].State.Voted = true
				nominated.State.VoteCount += 1
				msgPlayer += fmt.Sprintf("决意投给 [%s] \n", nominated.Name)
				game.Players[i].Log += msgPlayer
			}
			// 管家的投票在endVoting结算
			if game.Players[i].Character != Butler {
				msgAll += fmt.Sprintf("投票 [%s] 成功\n", nominated.Name)
			}
			// 总日志加入票池
			game.VotePool[player.Id] = msgAll
			// 发送个人日志
			emit(game, playerId)
			break
		}
	}
	// 投完票才会更新其他人的投标签，所以这里没有广播
}

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

	// 判断杀手
	if slayerTarget != nil {
		msgPlayer = "您"
		msgAll = ""
		for i, player := range game.Players {
			if player.Character == Slayer {
				// 不考虑酒鬼
				if player.State.Drunk {
					break
				}
				// 不考虑下毒了
				if player.State.Poisoned {
					game.Players[i].State.Bullet = false // 子弹不管怎样都会发射
					break
				}
				// 考虑子弹
				if game.Players[i].State.Bullet {
					msgAll += fmt.Sprintf("[%s] ", player.Name)
					game.Players[i].State.Bullet = false // 子弹不管怎样都会发射
					if slayerTarget.CharacterType == Demons || slayerTarget.State.RegardedAs == Imp {
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
						// 立即结算
						checkout(game, nil)
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
				}
			}
		}
	}
}

// 执行有顺序性，不可修改执行顺序
func checkoutNight(mux *sync.Mutex, game *model.Room) {
	mux.Lock()
	defer mux.Unlock()

	var msgPlayer = "您"
	var msgAll = ""

	// 承载技能释放者对象的池
	castPoolObj := map[*model.Player][]int{}
	for fromPlayerId, toPlayerIdSlice := range game.CastPool {
		for i, player := range game.Players {
			if player.Id == fromPlayerId {
				castPoolObj[&game.Players[i]] = []int{}
				for _, toPlayerId := range toPlayerIdSlice {
					for _, player := range game.Players {
						if player.Id == toPlayerId {
							castPoolObj[&game.Players[i]] = append(castPoolObj[&game.Players[i]], player.Index)
							break
						}
					}
				}
				break
			}
		}
	}

	// 各身份夜晚技能结算-有顺序
	// 判断下毒
	for fromPlayer, toPlayerIndexSlice := range castPoolObj {
		if fromPlayer.Character == Poisoner && !fromPlayer.State.Dead {
			game.Players[toPlayerIndexSlice[0]].State.Poisoned = true
			// 判断隐士被毒变非邪恶
			if game.Players[toPlayerIndexSlice[0]].Character == Recluse {
				game.Players[toPlayerIndexSlice[0]].State.Evil = false
				game.Players[toPlayerIndexSlice[0]].State.RegardedAs = Recluse
				if game.Players[toPlayerIndexSlice[0]].State.RegardedAsSaved == Imp {
					game.Players[toPlayerIndexSlice[0]].State.Demon = false
					// 判断隐士是否是占卜认定的恶魔
					var demonQuantity = 0
					var hasFortuneTeller = false
					for _, player := range game.Players {
						if player.State.Demon {
							demonQuantity += 1
						}
						if player.Character == FortuneTeller {
							hasFortuneTeller = true
						}
					}
					// 如果他是，那他demon还是true
					if hasFortuneTeller && demonQuantity == 2 {
						game.Players[toPlayerIndexSlice[0]].State.Demon = true
					}
				}
			}
			// 判断间谍被毒是邪恶
			if game.Players[toPlayerIndexSlice[0]].Character == Spy {
				game.Players[toPlayerIndexSlice[0]].State.Evil = true
			}
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
				var folk model.Player
				var other model.Player
				var character string
				if !player.State.Drunk && !player.State.Poisoned {
					// 生成随机信息
					for {
						randInt := rand.Intn(len(game.Players))
						if game.Players[randInt].CharacterType == Townsfolk && game.Players[randInt].Id != player.Id {
							folk = game.Players[randInt]
							character = game.Players[randInt].Character
							break
						}
					}
					for {
						randInt := rand.Intn(len(game.Players))
						if game.Players[randInt].Id != folk.Id && game.Players[randInt].Id != player.Id {
							other = game.Players[randInt]
							break
						}
					}
				} else {
					// 生成伪信息
					for {
						randInt := rand.Intn(len(game.Players))
						if game.Players[randInt].Id != player.Id {
							folk = game.Players[randInt]
							break
						}
					}
					for {
						randInt := rand.Intn(len(game.Players))
						if game.Players[randInt].Id != folk.Id && game.Players[randInt].Id != player.Id {
							other = game.Players[randInt]
							break
						}
					}
					// 生成伪村民身份名
					for {
						randInt := rand.Intn(len(TownsfolkPool))
						if TownsfolkPool[randInt] != Washerwoman && TownsfolkPool[randInt] != folk.Character && TownsfolkPool[randInt] != other.Character {
							character = TownsfolkPool[randInt]
							break
						}
					}
				}
				// 拼接日志
				msgAll += fmt.Sprintf("[%s] ", player.Name)
				var info string
				randInt := rand.Intn(2) // 随机顺序
				if randInt == 0 {
					info = fmt.Sprintf("发现 [%s] 和 [%s] 其中一个是 {%s}\n", folk.Name, other.Name, character)
				} else {
					info = fmt.Sprintf("发现 [%s] 和 [%s] 其中一个是 {%s}\n", other.Name, folk.Name, character)
				}
				msgPlayer += info
				msgAll += info
				game.Players[i].Log += msgPlayer
				game.Log += msgAll
				// 发送日志
				emit(game, player.Id)
			// 给图书管理员提供信息
			case Librarian:
				var outsider model.Player
				var other model.Player
				var hasOutsider bool
				var character string
				if !player.State.Drunk && !player.State.Poisoned {
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
							if (game.Players[randInt].CharacterType == Outsiders || game.Players[randInt].State.Drunk) && game.Players[randInt].Id != player.Id {
								outsider = game.Players[randInt]
								character = game.Players[randInt].Character
								break
							}
						}
						// 如果是酒鬼被选中
						if outsider.CharacterType != Outsiders {
							character = Drunk
						}
						for {
							randInt := rand.Intn(len(game.Players))
							if game.Players[randInt].Id != outsider.Id && game.Players[randInt].Id != player.Id {
								other = game.Players[randInt]
								break
							}
						}
					}
				} else {
					// 生成伪信息
					randInt := rand.Intn(len(game.Players))
					if game.Players[randInt].Id != player.Id {
						outsider = game.Players[randInt]
					}
					for {
						randInt := rand.Intn(len(game.Players))
						if game.Players[randInt].Id != outsider.Id && game.Players[randInt].Id != player.Id {
							other = game.Players[randInt]
							break
						}
					}
					// 生成伪外来者身份名
					for {
						randInt := rand.Intn(len(OutsidersPool))
						if OutsidersPool[randInt] != Recluse && OutsidersPool[randInt] != outsider.Character && OutsidersPool[randInt] != other.Character {
							character = OutsidersPool[randInt]
							break
						}
					}
				}
				// 拼接日志
				msgAll += fmt.Sprintf("[%s] ", player.Name)
				var info string
				if !hasOutsider {
					info = "发现本局 {没有外来者}\n"
				} else {
					randFixedNum := 51 // 假话：没有外来者，的概率是1/51
					if len(game.Players) > randFixedNum {
						randFixedNum = len(game.Players)
					}
					randInt := rand.Intn(randFixedNum)
					if randInt == 0 {
						info = "发现本局 {没有外来者}\n"
					} else if randInt%2 == 1 {
						info = fmt.Sprintf("发现 [%s] 和 [%s] 其中一个是 {%s}\n", outsider.Name, other.Name, character)
					} else {
						info = fmt.Sprintf("发现 [%s] 和 [%s] 其中一个是 {%s}\n", other.Name, outsider.Name, character)
					}
				}
				msgPlayer += info
				msgAll += info
				game.Players[i].Log += msgPlayer
				game.Log += msgAll
				// 发送日志
				emit(game, player.Id)
			// 给调查员提供信息
			case Investigator:
				var minion model.Player
				var other model.Player
				var character string
				if !player.State.Drunk && !player.State.Poisoned {
					// 生成随机信息
					for {
						randInt := rand.Intn(len(game.Players))
						if game.Players[randInt].CharacterType == Minions && game.Players[randInt].Id != player.Id {
							minion = game.Players[randInt]
							character = game.Players[randInt].Character
							break
						}
					}
					for {
						randInt := rand.Intn(len(game.Players))
						if game.Players[randInt].Id != minion.Id && game.Players[randInt].Id != player.Id {
							other = game.Players[randInt]
							break
						}
					}
				} else {
					// 生成伪信息
					for {
						randInt := rand.Intn(len(game.Players))
						if game.Players[randInt].Id != player.Id {
							minion = game.Players[randInt]
							break
						}
					}
					for {
						randInt := rand.Intn(len(game.Players))
						if game.Players[randInt].Id != minion.Id && game.Players[randInt].Id != player.Id {
							other = game.Players[randInt]
							break
						}
					}
					// 生成伪爪牙身份名
					for {
						randInt := rand.Intn(len(MinionsPool))
						if MinionsPool[randInt] != minion.Character && MinionsPool[randInt] != other.Character {
							character = MinionsPool[randInt]
							break
						}
					}
				}
				// 拼接日志
				msgAll += fmt.Sprintf("[%s] ", player.Name)
				var info string
				randInt := rand.Intn(2)
				if randInt == 0 {
					info = fmt.Sprintf("发现 [%s] 和 [%s] 其中一个是 {%s}\n", minion.Name, other.Name, character)
				} else {
					info = fmt.Sprintf("发现 [%s] 和 [%s] 其中一个是 {%s}\n", other.Name, minion.Name, character)
				}
				msgPlayer += info
				msgAll += info
				game.Players[i].Log += msgPlayer
				game.Log += msgAll
				// 发送日志
				emit(game, player.Id)
			// 给厨师提供信息
			case Chef:
				connected := 0 // 记录连座数
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
				// 生成伪信息
				if player.State.Drunk || player.State.Poisoned {
					evilQuantity := 0
					for _, player := range game.Players {
						if player.State.Evil {
							evilQuantity += 1
						}
					}
					for {
						randInt := rand.Intn(evilQuantity)
						if connected != randInt {
							connected = randInt
							break
						}
					}
				}
				// 拼接日志
				msgAll += fmt.Sprintf("[%s] ", player.Name)
				info := fmt.Sprintf("发现互为邻座的邪恶玩家有 { %d } 对\n", connected)
				msgPlayer += info
				msgAll += info
				game.Players[i].Log += msgPlayer
				game.Log += msgAll
				// 发送日志
				emit(game, player.Id)
			// 给共情者提供信息
			case Empath:
				evilQuantity := 0 // 记录左右邪恶玩家数量
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
				// 生成伪信息
				if player.State.Drunk || player.State.Poisoned {
					for {
						randInt := rand.Intn(3)
						if evilQuantity != randInt {
							evilQuantity = randInt
							break
						}
					}
				}
				// 拼接日志
				msgAll += fmt.Sprintf("[%s] ", player.Name)
				info := fmt.Sprintf("发现左右邻座的邪恶玩家有 { %d } 个\n", evilQuantity)
				msgPlayer += info
				msgAll += info
				game.Players[i].Log += msgPlayer
				game.Log += msgAll
				// 发送日志
				emit(game, player.Id)
			// 给间谍提供信息
			case Spy:
				// 拼接日志
				msgAll += fmt.Sprintf("[%s] 是间谍，知晓所有身份\n", player.Name)
				var info string
				info += "知晓所有身份：\n"
				for _, player := range game.Players {
					info += fmt.Sprintf("玩家 [%s] 的身份是 {%s}\n", player.Name, player.Character)
				}
				msgPlayer += info
				game.Players[i].Log += msgPlayer
				game.Log += msgAll
				// 发送日志
				emit(game, player.Id)
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
		// 判断士兵
		if game.Players[toPlayerIndexSlice[0]].Character == Soldier &&
			!game.Players[toPlayerIndexSlice[0]].State.Poisoned &&
			!game.Players[toPlayerIndexSlice[0]].State.Drunk {
			break
		}
		// 判断被僧侣守护
		if game.Players[toPlayerIndexSlice[0]].State.Protected &&
			!game.Players[toPlayerIndexSlice[0]].State.Poisoned {
			break
		}
		// 判断小恶魔被下毒
		if fromPlayer.Character == Imp && fromPlayer.State.Poisoned {
			break
		}
		// 必死人局面
		if fromPlayer.Character == Imp && !fromPlayer.State.Poisoned {
			// 死人状态改变在这个域的最后
			// 刀下毒者
			if game.Players[toPlayerIndexSlice[0]].Character == Poisoner {
				for i := range game.Players {
					game.Players[i].State.Poisoned = false
				}
			}
			// 自刀
			if game.Players[toPlayerIndexSlice[0]].Character == Imp {
				var scarletWoman *model.Player
				var minionsAlive []*model.Player
				var aliveQuantity int
				msgPlayer = "您"
				msgAll = ""
				for i, player := range game.Players {
					if !player.State.Dead {
						aliveQuantity += 1
					}
					if player.Character == ScarletWoman && !player.State.Dead {
						scarletWoman = &game.Players[i]
					} else {
						if player.CharacterType == Minions && !player.State.Dead && player.Character != ScarletWoman {
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
					// 拼接日志
					info := "已变为 {小恶魔} \n"
					msgPlayer += info
					msgAll += fmt.Sprintf("[%s] ", scarletWoman.Name) + info
					scarletWoman.Log += msgPlayer
					game.Log += msgAll
					// 发送日志
					emit(game, scarletWoman.Id)
				}
				// 如果没有魅魔或魅魔死了
				if reflect.ValueOf(scarletWoman).IsZero() && len(minionsAlive) != 0 {
					randInt := rand.Intn(len(minionsAlive))
					minionsAlive[randInt].CharacterType = Demons
					minionsAlive[randInt].Character = Imp
					minionsAlive[randInt].State.Evil = true
					minionsAlive[randInt].State.Demon = true
					// 下毒转变为恶魔，当夜技能失效
					if minionsAlive[randInt].Character == Poisoner {
						for i := range game.Players {
							game.Players[i].State.Poisoned = false
						}
					}
					// 拼接日志
					info := "已变为小恶魔\n"
					msgPlayer += info
					msgAll += fmt.Sprintf("[%s] ", minionsAlive[randInt].Name) + info
					minionsAlive[randInt].Log += msgPlayer
					game.Log += msgAll
					// 发送日志
					emit(game, minionsAlive[randInt].Id)
				}
			}
			// 判断刀市长
			if game.Players[toPlayerIndexSlice[0]].Character == Mayor &&
				!game.Players[toPlayerIndexSlice[0]].State.Poisoned &&
				!game.Players[toPlayerIndexSlice[0]].State.Drunk {
				for {
					randInt := rand.Intn(len(game.Players))
					if !game.Players[randInt].State.Dead &&
						game.Players[randInt].CharacterType != Demons {
						// 死的是除了恶魔的其他任意一人
						game.Players[randInt].State.Dead = true
						game.Players[randInt].Ready.Nominate = false
						game.Players[randInt].Ready.Nominated = false
						killed = &game.Players[randInt]
						break
					}
				}
				break
			}
			// 死的人
			game.Players[toPlayerIndexSlice[0]].State.Dead = true
			game.Players[toPlayerIndexSlice[0]].Ready.Nominate = false
			game.Players[toPlayerIndexSlice[0]].Ready.Nominated = false
			killed = &game.Players[toPlayerIndexSlice[0]]
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
				info := ""
				// 没有死人
				if killed == nil {
					break
				}
				// 不是酒鬼，没被毒或被守护，死的正是守鸦人自己
				msgAll += fmt.Sprintf("[%s] ", player.Name)
				for fromPlayer, toPlayerIndexSlice := range castPoolObj {
					if fromPlayer.Id == player.Id {
						character := ""
						// 给守鸦人提供正确信息
						if !player.State.Drunk && !player.State.Poisoned && player.Id == killed.Id {
							// 看隐士情况，是看成他被当成的身份
							if game.Players[toPlayerIndexSlice[0]].Character == Recluse {
								character = game.Players[toPlayerIndexSlice[0]].State.RegardedAs
							} else {
								character = game.Players[toPlayerIndexSlice[0]].Character
							}
							info = fmt.Sprintf("发现 [%s] 的身份是 {%s}\n", game.Players[toPlayerIndexSlice[0]].Name, character)
							break
						}
						// 给守鸦人提供伪信息
						if (player.State.Drunk || player.State.Poisoned) && player.Id == killed.Id {
							for {
								randInt := rand.Intn(len(AllChars))
								if AllChars[randInt] != Ravenkeeper {
									character = AllChars[randInt]
									break
								}
							}
							info = fmt.Sprintf("发现 [%s] 的身份是 {%s}\n", game.Players[toPlayerIndexSlice[0]].Name, character)
							break
						}
					}
				}
				msgPlayer += info
				msgAll += info
				game.Players[player.Index].Log += msgPlayer
				game.Log += msgAll
				// 发送日志
				emit(game, player.Id)
			// 给共情者提供信息
			case Empath:
				if !player.State.Dead { // 当晚死亡得不到信息
					evilQuantity := 0 // 记录左右邪恶玩家数量
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
					var leftPrev = -1
					var rightPrev = -1
					for {
						if leftPrev != left {
							leftPrev = left
							if game.Players[left].State.Dead {
								left--
								if left < 0 {
									left = len(game.Players) - 1
								}
							} else {
								if game.Players[left].State.Evil {
									evilQuantity += 1
								}
							}
						}
						if rightPrev != right {
							rightPrev = right
							if game.Players[right].State.Dead {
								right++
								if right > len(game.Players)-1 {
									right = 0
								}
							} else {
								if game.Players[right].State.Evil {
									evilQuantity += 1
								}
							}
						}
						if leftPrev == left && rightPrev == right || left == right {
							break
						}
					}
					// 生成伪信息
					if player.State.Drunk || player.State.Poisoned {
						for {
							randInt := rand.Intn(3)
							if randInt != evilQuantity {
								evilQuantity = randInt
								break
							}
						}
					}
					// 拼接日志
					msgAll += fmt.Sprintf("[%s] ", player.Name)
					info := fmt.Sprintf("发现左右邻座的邪恶玩家有 { %d } 个\n", evilQuantity)
					msgPlayer += info
					msgAll += info
					game.Players[i].Log += msgPlayer
					game.Log += msgAll
					// 发送日志
					emit(game, player.Id)
				}
			// 给掘墓人提供信息
			case Undertaker:
				if !player.State.Dead {
					if game.Executed == nil {
						// 无人被处决
						msgAll += fmt.Sprintf("[%s] ", player.Name)
						info := "发现昨日无人被处决\n"
						msgPlayer += info
						msgAll += info
					} else {
						character := ""
						if !player.State.Drunk && !player.State.Poisoned {
							// 生成死亡玩家身份信息
							character = game.Executed.Character
							// 看隐士情况，是看成他被当成的身份
							if game.Executed.Character == Recluse {
								character = game.Executed.State.RegardedAs
							}
						} else {
							// 生成伪信息
							for {
								randInt := rand.Intn(len(AllChars))
								if AllChars[randInt] != player.Character && AllChars[randInt] != game.Executed.Character {
									character = AllChars[randInt]
									break
								}
							}
						}
						// 拼接日志
						msgAll += fmt.Sprintf("[%s] ", player.Name)
						info := fmt.Sprintf("发现今晚被处决的玩家 [%s] 的身份是 {%s}\n", game.Executed.Name, character)
						msgPlayer += info
						msgAll += info
					}
					game.Players[i].Log += msgPlayer
					game.Log += msgAll
					// 发送日志
					emit(game, player.Id)
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
			emit(game, fromPlayer.Id)
			break
		}
	}
	// 给管家提供信息
	for fromPlayer, toPlayerIndexSlice := range castPoolObj {
		if fromPlayer.Character == Butler && !fromPlayer.State.Dead {
			if !fromPlayer.State.Poisoned {
				game.Players[toPlayerIndexSlice[0]].State.Master = true
			}
			msgPlayer = "您"
			msgAll = ""
			// 拼接日志
			msgAll += fmt.Sprintf("[%s] ", fromPlayer.Name)
			info := fmt.Sprintf("认定 [%s] 为主人，他投你可选投，他不投你也不能投\n", game.Players[toPlayerIndexSlice[0]].Name)
			msgPlayer += info
			msgAll += info
			game.Players[fromPlayer.Index].Log += msgPlayer
			game.Log += msgAll
			// 发送日志
			emit(game, fromPlayer.Id)
			break
		}
	}

	// ------- 结算 -------
	msg := ""
	// 结算本阶段
	if killed == nil {
		msg += "今夜是 平安夜\n"
	} else {
		msg += fmt.Sprintf("今夜 [%s] 死亡\n", killed.Name)
	}
	// 拼接日志
	for i := range game.Players {
		game.Players[i].Log += msg
	}
	game.Log += msg
	// 发送日志
	broadcast(game)
	// 结算本局
	checkout(game, nil) // 这里不要传入executed，因为晚上不处决人，晚上可能死圣徒
}

func checkoutDay(mux *sync.Mutex, game *model.Room) {
	mux.Lock()
	defer mux.Unlock()
	// 结算本局
	checkout(game, game.Executed)
}

// checkout 结算本局
func checkout(game *model.Room, executed *model.Player) {
	msg := ""
	var realDemonCount int     // 恶魔数量，被占卜认定的不算
	var hasSlayerBullet bool   // 有杀手且杀手有子弹
	var aliveCount int         // 活人数量
	var canVote int            // 可投票数量
	var evilAliveCount int     // 邪恶玩家存活数量
	var mayorAlive bool        // 市长是否存活
	var scarletWomanAlive bool // 魅魔是否存活
	for _, player := range game.Players {
		// 对应平民胜利条件1
		if player.CharacterType == Demons && !player.State.Dead {
			realDemonCount++
		}
		if player.Character == ScarletWoman && !player.State.Dead && !player.State.Poisoned {
			scarletWomanAlive = true
		}
		// 对应平民胜利条件2
		if player.Character == Mayor && !player.State.Dead && !player.State.Drunk && !player.State.Poisoned {
			mayorAlive = true
		}
		// 可投票数 = 持票死人人数 + 活人人数
		if !player.State.Dead {
			canVote += 1
		} else {
			canVote += player.Ready.Vote
		}
		// 对应邪恶胜利条件2
		if !player.State.Dead {
			aliveCount++
		}
		// 对应邪恶胜利条件3
		if player.Character == Slayer && player.State.Bullet {
			hasSlayerBullet = true
		}
		if !player.State.Dead && (player.CharacterType == Demons || player.CharacterType == Minions) {
			evilAliveCount++
		}
	}
	// 平民胜利条件1（恶魔受不了了自杀情况），这里有三种铲除恶魔的可能：1、杀手，2、处决，3、自刀。
	// 杀手和处决在此处判断魅魔，自刀在判定刀人时已经发生转化，所以realDemonCount不可能为0
	if game.Result == "" && realDemonCount == 0 && (!scarletWomanAlive || scarletWomanAlive && aliveCount >= 5) {
		msg += "达成平民胜利条件一：恶魔被铲除干净\n"
		msg += "本局结束，平民胜利\n"
		game.Result = "平民阵营胜利"
	}
	// 平民胜利条件2
	if game.Result == "" && aliveCount == 3 && !game.State.Night && mayorAlive && executed == nil {
		msg += "达成平民胜利条件二：白天剩三人，其中一个是市长，且当日无人被处决\n"
		msg += "本局结束，平民胜利\n"
		game.Result = "平民阵营胜利"
	}
	// 邪恶胜利条件1
	if game.Result == "" && executed != nil && executed.Character == Saint && !executed.State.Poisoned {
		msg += "达成邪恶胜利条件一：圣徒被投票处决，且未中毒\n"
		msg += "本局结束，邪恶胜利\n"
		game.Result = "邪恶阵营胜利"
	}
	// 邪恶胜利条件2
	if game.Result == "" && evilAliveCount == aliveCount {
		msg += "达成邪恶胜利条件二：平民阵营被屠城\n"
		msg += "本局结束，邪恶胜利\n"
		game.Result = "邪恶阵营胜利"
	}
	// 邪恶胜利条件3
	halfAlive := int(math.Ceil(float64(aliveCount / 2)))
	if game.Result == "" && aliveCount <= 4 && canVote-evilAliveCount <= halfAlive && evilAliveCount >= halfAlive && !hasSlayerBullet && !mayorAlive {
		msg += "达成邪恶胜利条件三：活人数小于等于4，平民可投的票数不大于活人的半数（向上取整），且存活的邪恶玩家数量不小于活人的半数（向上取整），且没有杀手或有杀手没有子弹，且没有市长或市长已死或酒鬼市长\n"
		msg += "本局结束，邪恶胜利\n"
		game.Result = "邪恶阵营胜利"
	}

	if game.Result != "" {
		// 拼接日志
		for i := range game.Players {
			game.Players[i].Log += msg
		}
		game.Log += msg
		game.Status = "复盘中"
		// 发送game 以便前端跳转review
		broadcast(game)
	}
}

func findThreeCharactersNotInGame(players []model.Player) string {
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
