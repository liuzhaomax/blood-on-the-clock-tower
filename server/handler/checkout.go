package handler

import (
	"fmt"
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/model"
	"math"
	"math/rand"
	"reflect"
	"sync"
)

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
					// 判断隐士是否是占卜认定的恶魔
					// 这个恶魔数量至少为1，有占卜或恶魔隐士就是2，有占卜且有恶魔隐士且恶魔隐士是占卜认定的恶魔是2，有占卜且有恶魔隐士就是3
					// 因为这是在恶魔隐士的域内，所以没有可能是1，也没有可能出现有占卜没有恶魔隐士的情况
					// 情况化简为：2 -> 没占卜有恶魔隐士 || 有占卜且有恶魔隐士且恶魔隐士是占卜认定的恶魔-重叠情况，3 -> 有占卜且有恶魔隐士，且认定恶魔与恶魔隐士不重叠
					// 进一步化简为：有占卜 && 恶魔数量是2 -> 重叠
					var demonQuantity = 0        // 这个恶魔数量，是不看死活的
					var hasFortuneTeller = false // 是否有占卜师
					for _, player := range game.Players {
						if player.State.Demon {
							demonQuantity += 1
						}
						if player.Character == FortuneTeller {
							hasFortuneTeller = true
						}
					}
					// 先把他变成不是恶魔，再判断
					game.Players[toPlayerIndexSlice[0]].State.Demon = false
					// 如果有占卜且有恶魔隐士且恶魔隐士是占卜认定的恶魔，那他demon还是true
					if hasFortuneTeller && demonQuantity == 2 {
						game.Players[toPlayerIndexSlice[0]].State.Demon = true
					}
				}
			}
			// 判断间谍被毒是邪恶
			if game.Players[toPlayerIndexSlice[0]].Character == Spy {
				game.Players[toPlayerIndexSlice[0]].State.Evil = true
				game.Players[toPlayerIndexSlice[0]].State.RegisteredAsType = Minions
				game.Players[toPlayerIndexSlice[0]].State.RegardedAs = Spy
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
						if game.Players[randInt].Character == Spy && game.Players[randInt].State.RegisteredAsType == Townsfolk {
							folk = game.Players[randInt]
							character = game.Players[randInt].State.RegardedAs
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
				for _, player := range game.Players {
					if player.CharacterType == Outsiders && player.Character != Recluse ||
						(player.Character == Spy && player.State.RegisteredAsType == Outsiders && player.State.RegardedAs != Recluse) ||
						player.State.Drunk {
						hasOutsider = true
						break
					}
				}
				if !player.State.Drunk && !player.State.Poisoned {
					if hasOutsider {
						// 生成随机信息
						for {
							randInt := rand.Intn(len(game.Players))
							if game.Players[randInt].CharacterType == Outsiders && game.Players[randInt].Id != player.Id &&
								game.Players[randInt].Character != Recluse && game.Players[randInt].Character != Spy {
								outsider = game.Players[randInt]
								character = game.Players[randInt].Character
								break
							}
							// 如果是酒鬼被选中
							if game.Players[randInt].State.Drunk && game.Players[randInt].Id != player.Id &&
								game.Players[randInt].Character != Recluse && game.Players[randInt].Character != Spy {
								outsider = game.Players[randInt]
								character = Drunk
								break
							}
							// 如果间谍被选中
							if game.Players[randInt].Character == Spy && game.Players[randInt].State.RegisteredAsType == Outsiders {
								outsider = game.Players[randInt]
								character = game.Players[randInt].State.RegardedAs
								break
							}
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
				var hasSpyOnly bool
				var hasSpy bool
				var minionsQuantity int
				for _, player := range game.Players {
					if player.CharacterType == Minions {
						minionsQuantity += 1
					}
					if player.Character == Spy && !player.State.Poisoned {
						hasSpy = true
					}
				}
				// 判断是否只有一个爪牙，是间谍，且没中毒
				if minionsQuantity == 1 && hasSpy {
					hasSpyOnly = true
				}
				if !player.State.Drunk && !player.State.Poisoned {
					if !hasSpyOnly { // 如果只有间谍一个爪牙，下面的函数会陷入死循环
						// 生成随机信息
						for {
							randInt := rand.Intn(len(game.Players))
							if game.Players[randInt].CharacterType == Minions && game.Players[randInt].Id != player.Id && game.Players[randInt].Character != Spy {
								minion = game.Players[randInt]
								character = game.Players[randInt].Character
								break
							}
							if game.Players[randInt].Character == Recluse &&
								(game.Players[randInt].State.RegardedAs == Spy || game.Players[randInt].State.RegardedAs == Poisoner ||
									game.Players[randInt].State.RegardedAs == ScarletWoman || game.Players[randInt].State.RegardedAs == Baron) {
								minion = game.Players[randInt]
								character = game.Players[randInt].State.RegardedAs
								break
							}
							if game.Players[randInt].Character == Spy && game.Players[randInt].State.Poisoned {
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
				if hasSpyOnly {
					info = "发现本局 {没有爪牙}\n"
				} else {
					randInt := rand.Intn(2)
					if randInt == 0 {
						info = fmt.Sprintf("发现 [%s] 和 [%s] 其中一个是 {%s}\n", minion.Name, other.Name, character)
					} else {
						info = fmt.Sprintf("发现 [%s] 和 [%s] 其中一个是 {%s}\n", other.Name, minion.Name, character)
					}
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
		if fromPlayer.Character == Imp && !fromPlayer.State.Dead && !fromPlayer.State.Poisoned {
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
			// 必死人局面
			if fromPlayer.Character == Imp && !fromPlayer.State.Poisoned {
				// 死人状态改变在这个域的最后
				// 刀下毒者
				if game.Players[toPlayerIndexSlice[0]].Character == Poisoner {
					for i := range game.Players {
						game.Players[i].State.Poisoned = false
					}
				}
				// 刀圣女
				if game.Players[toPlayerIndexSlice[0]].Character == Virgin {
					game.Players[toPlayerIndexSlice[0]].State.Blessed = false
				}
				// 刀杀手
				if game.Players[toPlayerIndexSlice[0]].Character == Slayer {
					game.Players[toPlayerIndexSlice[0]].State.Bullet = false
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
							// 看隐士和间谍情况，是看成他被当成的身份
							if game.Players[toPlayerIndexSlice[0]].Character == Recluse || game.Players[toPlayerIndexSlice[0]].Character == Spy {
								character = game.Players[toPlayerIndexSlice[0]].State.RegardedAs
							} else {
								character = game.Players[toPlayerIndexSlice[0]].Character
							}
							// 可以看到酒鬼
							if game.Players[toPlayerIndexSlice[0]].State.Drunk {
								character = Drunk
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
							// 看隐士和间谍情况，是看成他被当成的身份
							if game.Executed.Character == Recluse || game.Executed.Character == Spy {
								character = game.Executed.State.RegardedAs
							}
							// 可以看到酒鬼
							if game.Executed.State.Drunk {
								character = Drunk
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
			info := fmt.Sprintf("认定 [%s] 为主人，他投你可选投，他不投你票无效\n", game.Players[toPlayerIndexSlice[0]].Name)
			msgPlayer += info
			msgAll += info
			game.Players[fromPlayer.Index].Log += msgPlayer
			game.Log += msgAll
			// 发送日志
			emit(game, fromPlayer.Id)
			break
		}
	}
	// 给间谍提供魔典
	for i, player := range game.Players {
		if player.Character == Spy && !player.State.Poisoned && !player.State.Dead {
			game.Players[i].Log = game.Log
			// 发送日志
			emit(game, player.Id)
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
	checkout(game, game.Executed)
}

func checkoutDay(mux *sync.Mutex, game *model.Room) {
	mux.Lock()
	defer mux.Unlock()
	// 结算处决
	execute(game)
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
	var poisonerAlive bool     // 下毒者是否存活
	var demonCount int         // 恶魔数量（不论死活）
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
		if player.Character == Poisoner && !player.State.Dead {
			poisonerAlive = true
		}
		// 可投票数
		if player.Ready.Vote {
			canVote += 1
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
		if player.CharacterType == Demons {
			demonCount++
		}
	}
	// 平民胜利条件1（恶魔受不了了自杀情况），这里有三种铲除恶魔的可能：1、杀手，2、处决，3、自刀。
	// 处决在结算投票时判定，枪杀在枪手施法后判定，自刀在判定刀人时判定，所以realDemonCount不可能为0
	// 魅魔再判定为双保险可删没测
	if game.Result == "" && realDemonCount == 0 && (!scarletWomanAlive || scarletWomanAlive && aliveCount < 5) {
		msg += "达成平民胜利条件一：恶魔被铲除干净\n"
		msg += "本局结束，平民胜利\n"
		game.Result = "平民阵营胜利"
	}
	// 平民胜利条件2
	if game.Result == "" && aliveCount == 3 && mayorAlive && executed == nil && !game.State.Night && !poisonerAlive {
		msg += "达成平民胜利条件二：白天剩三人，其中一个是市长，且当日无人被处决，且三人不是下毒者市长小恶魔的组合\n"
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
	if game.Result == "" && aliveCount <= 4 && demonCount == 1 && canVote-evilAliveCount <= halfAlive && evilAliveCount >= halfAlive && !hasSlayerBullet && !mayorAlive {
		msg += "达成邪恶胜利条件三：活人数小于等于4，未发生爪牙转化为恶魔，平民可投的票数不大于活人的半数（向上取整），且存活的邪恶玩家数量不小于活人的半数（向上取整），且没有杀手或有杀手没有子弹，且没有市长或市长已死或酒鬼市长\n"
		msg += "本局结束，邪恶胜利\n"
		game.Result = "邪恶阵营胜利"
	}

	if game.Result != "" {
		// 拼接日志
		for i := range game.Players {
			game.Players[i].Log += msg
		}
		game.Log += msg
		game.Status = RoomReviewing
		// 发送game 以便前端跳转review
		broadcast(game)
	}
}
