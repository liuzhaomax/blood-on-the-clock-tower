package handler

import (
	"fmt"
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/model"
	"sync"
)

func endVoting(mux *sync.Mutex, game *model.Room) {
	mux.Lock()
	defer mux.Unlock()

	if !game.State.VotingStep {
		return
	}

	var msg string

	// 判断管家
	var butlerPlayer *model.Player // 管家玩家
	for i, player := range game.Players {
		if player.Character == Butler {
			butlerPlayer = &game.Players[i]
			break
		}
	}
	// 主人没投，管家投了，票数减一，因为如果主人不投票，则管家不能跟票
	if !game.Nominated.State.VotedFromMaster && game.Nominated.State.VotedFromButler {
		game.VotePool[game.Nominated.Id] -= 1
		// 拼接日志
		msgButler := "主人未投票，您"
		msgAll := fmt.Sprintf("主人未投票，[%s] ", butlerPlayer.Name)
		info := fmt.Sprintf("投给 [%s] 的票无效！\n", game.Nominated.Name)
		msgButler += info
		msgAll += info
		butlerPlayer.Log += msgButler
		game.Log += msgAll
		// 发送日志 - 告诉管家投票无效
		emit(game, butlerPlayer.Id)
	}
	// 主人投了，管家也投了，要加入管家投票成功日志，对应vote函数的管家逻辑
	if game.Nominated.State.VotedFromMaster && game.Nominated.State.VotedFromButler {
		msg += fmt.Sprintf("[%s] 投票 [%s] 成功\n", butlerPlayer.Name, game.Nominated.Name)
	}
	// 打印所有投票成功的票型
	msg += game.VoteLogs[game.Nominated.Id]
	for i := range game.Players {
		game.Players[i].Log += msg
	}
	game.Log += msg
	// 退出投票处决环节
	game.State.VotingStep = false
	// 发送日志
	broadcast(game)
}
