package handler

import (
	"fmt"
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/model"
	"sync"
)

func vote(mux *sync.Mutex, game *model.Room, playerId string) {
	mux.Lock()
	defer mux.Unlock()

	var msgAll = ""
	var msgPlayer = "您"

	for i, player := range game.Players {
		if player.Id == playerId && player.Ready.Vote && game.State.VotingStep {
			msgAll += fmt.Sprintf("[%s] ", player.Name)
			if game.Nominated != nil && game.Players[i].Ready.Vote {
				if player.State.Dead {
					game.Players[i].Ready.Vote = false // 死人投了就不能再投票了
				}
				game.VotePool[game.Nominated.Id] += 1
				msgPlayer += fmt.Sprintf("决意投给 [%s] \n", game.Nominated.Name)
				game.Players[i].Log += msgPlayer
				// 管家的投票在endVoting结算
				if game.Players[i].Character != Butler {
					msgAll += fmt.Sprintf("投票 [%s] 成功\n", game.Nominated.Name)
					// 总日志加入票池
					game.VoteLogs[game.Nominated.Id] += msgAll
				}
				if game.Players[i].Character == Butler {
					game.Nominated.State.VotedFromButler = true
				}
				if game.Players[i].State.Master {
					game.Nominated.State.VotedFromMaster = true
				}
				// 发送个人日志
				emit(game, playerId)
				break
			}
		}
	}
}
