package handler

import (
	"fmt"
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/model"
	"math"
)

func execute(game *model.Room) {
	var msg string
	var executeeVoteCount int
	game.Executed, executeeVoteCount = findExecutee(game)
	if game.Executed != nil {
		game.Executed.State.Dead = true
		game.Executed.Ready.Nominate = false
		game.Executed.Ready.Nominated = false
		msg += fmt.Sprintf("处决结果：[%s] 获得 %d 票，被处决，死亡\n", game.Executed.Name, executeeVoteCount)
	} else {
		msg += "处决结果：无人被处决\n"
	}
	for i := range game.Players {
		game.Players[i].Log += msg
	}
	game.Log += msg

	// 判断魅魔 - 有人被处决，处决的人是小恶魔，活人大于等于5个，有魅魔且没死
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
	if game.Executed != nil && game.Executed.Character == Imp && aliveCount-1 >= 5 &&
		scarletWomanIndex > 0 && !game.Players[scarletWomanIndex].State.Poisoned { // 活人人数是转换死人状态之前统计的，要减一
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

	// 发送日志
	broadcast(game)
	// 立即结算
	checkout(game, game.Executed)
}

func findExecutee(game *model.Room) (*model.Player, int) {
	// 被圣女弹死
	if game.Executed != nil {
		return game.Executed, 0
	}
	// 无人被提名
	if len(game.VotePool) == 0 {
		return nil, 0
	}
	var aliveCount int // 活人数量
	for _, player := range game.Players {
		if !player.State.Dead {
			aliveCount++
		}
	}
	var halfAliveCount = int(math.Ceil(float64(aliveCount) / 2))
	var executeeId string
	var executeeVoteCount int
	var isHighestRepeated bool
	for nominatedId, voteCount := range game.VotePool {
		if voteCount >= halfAliveCount && voteCount == executeeVoteCount {
			isHighestRepeated = true
			continue
		}
		if voteCount >= halfAliveCount && voteCount > executeeVoteCount {
			executeeId = nominatedId
			executeeVoteCount = voteCount
		}
	}
	if executeeId != "" && !isHighestRepeated {
		for i, player := range game.Players {
			if player.Id == executeeId {
				return &game.Players[i], executeeVoteCount
			}
		}
	}
	return nil, executeeVoteCount
}
