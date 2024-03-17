package handler

import (
	"fmt"
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/model"
	"sync"
)

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
				if targets[0] == player.Id && player.Ready.Nominated { // 死了也能被提名
					game.Players[i].Ready.Nominate = false  // 发动提名者不能再提名
					game.Players[j].Ready.Nominated = false // 被提名者不能再被提名
					game.Nominated = &game.Players[j]
					game.VotePool[game.Nominated.Id] = 0
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
	var virgin *model.Player
	for i, player := range game.Players {
		if player.Character == Virgin && player.Id == targets[0] && player.State.Blessed && !player.State.Drunk && !player.State.Dead {
			virgin = &game.Players[i]
			virgin.State.Blessed = false
			msg = msgName
			for i, player := range game.Players {
				if player.Id == playerId && !player.State.Drunk && !player.State.Dead && !virgin.State.Poisoned &&
					(player.CharacterType == Townsfolk ||
						(player.Character == Spy && player.State.RegisteredAsType == Townsfolk && !player.State.Poisoned)) {
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
	// 进入投票环节 圣女条件成立就不用进投票环节
	if canGoToVotingStep {
		game.State.VotingStep = true
		// 广播votingStep
		broadcast(game)
	}
}
