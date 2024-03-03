package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/model"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

func LoadGame(w http.ResponseWriter, r *http.Request) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	cfg := model.GetConfig()

	// 获取URL参数
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.NotFound(w, r)
		return
	}
	roomId := parts[2]
	playerId := parts[3]

	CfgMutex.Lock()
	room, _ := findRoom(roomId)
	// 推入game连接池
	if cfg.GameConnPool[roomId] == nil {
		cfg.GameConnPool[roomId] = map[string]*websocket.Conn{}
	}
	cfg.GameConnPool[roomId][playerId] = conn
	// 初始化game的锁
	if cfg.MuxPool[roomId] == nil {
		cfg.MuxPool[roomId] = &sync.RWMutex{}
	}
	CfgMutex.Unlock()

	mux := cfg.MuxPool[roomId]

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway) {
				log.Println("Client disconnected:", err)
				return
			}
			log.Println("Read error Game:", err)
			return
		}

		action := string(p)
		switch action {
		case "load_game":
			initGame(mux, room)
		case "quit_game":
			quitGame(mux, room, playerId)
		}

		// 检测是否房间内所有人都退出游戏
		detectIfAllQuited(mux, room)

		time.Sleep(time.Millisecond * 50)
	}
}

func initGame(mux *sync.RWMutex, room *model.Room) {
	cfg := model.GetConfig()
	mux.Lock()
	defer mux.Unlock()
	if room == nil {
		return
	}

	if !room.Init {
		// 初始化
		room.Init = true
		room.CreatedAt = time.Now().Format(time.RFC3339)
		room.Result = ""
		room.Log = ""
		room.CastPool = map[string][]string{}
		room.VotePool = map[string]string{}
		room.State = model.GameState{}
		// 初始化玩家状态 防止非法返回房间引起bug
		for i, player := range room.Players {
			newPlayer := model.Player{}
			newPlayer.Id = player.Id
			newPlayer.Name = player.Name
			newPlayer.Index = player.Index
			room.Players[i] = newPlayer
		}
		// 分配身份
		var replaceDrunk string
		if room.Players[0].Character == "" {
			room.Players, replaceDrunk = allocateCharacter(room.Players)
		}
		// 初始化玩家状态 依赖身份
		room.Players = initStatus(room.Players, replaceDrunk)
		// 保存玩家身份到总日志
		var hasFortuneTeller bool
		var hasRecluse bool
		room.Log = "本局配置：\n"
		for _, player := range room.Players {
			room.Log += fmt.Sprintf("玩家 [%s] 的身份是 {%s} \n", player.Name, player.Character)
			if player.State.Drunk {
				room.Log += fmt.Sprintf("玩家 [%s] 的身份其实是 {%s} ~\n", player.Name, Drunk)
			}
			if player.Character == Recluse {
				room.Log += fmt.Sprintf("玩家 [%s] 的被当作的身份是 {%s} ~\n", player.Name, player.State.RegardedAs)
				hasRecluse = true
			}
			if player.Character == FortuneTeller {
				hasFortuneTeller = true
			}
		}
		if hasFortuneTeller && !hasRecluse {
			for _, player := range room.Players {
				if player.State.Demon && player.CharacterType != Demons {
					room.Log += fmt.Sprintf("玩家 [%s] 是占卜师认定的恶魔~\n", player.Name)
					break
				}
			}
		}
		if hasFortuneTeller && hasRecluse {
			var extraDemonQuantity int
			for _, player := range room.Players {
				if player.State.Demon && player.CharacterType != Demons {
					extraDemonQuantity += 1
				}
			}
			if extraDemonQuantity == 2 {
				for _, player := range room.Players {
					if player.Character == Recluse {
						room.Log += fmt.Sprintf("玩家 [%s] 是占卜师认定的恶魔~\n", player.Name)
						break
					}
				}
			}
			if extraDemonQuantity == 3 {
				for _, player := range room.Players {
					if player.Character != Recluse && player.CharacterType != Demons && player.State.Demon {
						room.Log += fmt.Sprintf("玩家 [%s] 是占卜师认定的恶魔~\n", player.Name)
						break
					}
				}
			}
		}
		room.Log += "------------本--局--开--始------------\n"
	}

	// 群发game
	if room.Result == "" {
		marshaledRoom, err := json.Marshal(room)
		if err != nil {
			log.Println("JSON marshal error:", err)
			return
		}
		for _, conn := range cfg.GameConnPool[room.Id] {
			if err = conn.WriteMessage(websocket.TextMessage, marshaledRoom); err != nil {
				log.Println("Write error:", err)
				return
			}
		}
	}
}

func allocateCharacter(players []model.Player) ([]model.Player, string) {
	playerNum := len(players)
	var characterNumSlice []int
	// 索引0是村民，索引1是外来者，索引2是爪牙，索引3是恶魔
	switch playerNum {
	case 5:
		characterNumSlice = []int{3, 0, 1, 1}
	case 6:
		characterNumSlice = []int{3, 1, 1, 1}
	case 7:
		characterNumSlice = []int{5, 0, 1, 1}
	case 8:
		characterNumSlice = []int{5, 1, 1, 1}
	case 9:
		characterNumSlice = []int{5, 2, 1, 1}
	case 10:
		characterNumSlice = []int{7, 0, 2, 1}
	case 11:
		characterNumSlice = []int{7, 1, 2, 1}
	case 12:
		characterNumSlice = []int{7, 2, 2, 1}
	case 13:
		characterNumSlice = []int{9, 0, 3, 1}
	case 14:
		characterNumSlice = []int{9, 1, 3, 1}
	case 15:
		characterNumSlice = []int{9, 2, 3, 1}
	default:
		characterNumSlice = []int{5, 2, 1, 1} // 9人
	}
	// 里面存放的是针对对应身份类型池的index
	// 这些index的数量取决于characterNumSlice中的对应位置的数字，比如本局有5个村民，那townsfolkRandNums的长度就是5
	// 这些index都是根据对应的身份类型池的总数量的随机数，比如本局有3个村民，[0,7,4]，对应到TownsfolkPool的0/7/4索引位置
	townsfolkRandNums := []int{}
	outsidersRandNums := []int{}
	minionsRandNums := []int{}
	demonsRandNums := []int{}
	for i, num := range characterNumSlice {
		if num != 0 {
			switch i {
			case 0:
				townsfolkRandNums = genRandomPositionSlice(townsfolkRandNums, TownsfolkPool, num)
			case 1:
				outsidersRandNums = genRandomPositionSlice(outsidersRandNums, OutsidersPool, num)
			case 2:
				minionsRandNums = genRandomPositionSlice(minionsRandNums, MinionsPool, num)
			case 3:
				demonsRandNums = genRandomPositionSlice(demonsRandNums, DemonsPool, num)
			}
		}
	}

	var characterPoolForSelection []string
	var replaceDrunk string
	var repeatFlag bool
	for _, randIdx := range demonsRandNums {
		characterPoolForSelection = append(characterPoolForSelection, DemonsPool[randIdx])
	}
	for _, randIdx := range minionsRandNums {
		characterPoolForSelection = append(characterPoolForSelection, MinionsPool[randIdx])
		// 处理男爵
		if MinionsPool[randIdx] == Baron {
			outsidersNumsLength := len(outsidersRandNums)
			for {
				if len(outsidersRandNums) == outsidersNumsLength+2 {
					break
				}
				repeatFlag = false
				randIdxOutsiders := rand.Intn(len(OutsidersPool))
				for _, num := range outsidersRandNums {
					if randIdxOutsiders == num {
						repeatFlag = true
						break
					}
				}
				if !repeatFlag {
					outsidersRandNums = append(outsidersRandNums, randIdxOutsiders)
				}
			}
			townsfolkRandNums = townsfolkRandNums[:len(townsfolkRandNums)-2]
		}
	}
	if len(outsidersRandNums) != 0 {
		for _, randIdx := range outsidersRandNums {
			characterPoolForSelection = append(characterPoolForSelection, OutsidersPool[randIdx])
			// 处理酒鬼 酒鬼还在池中，后续阶段由replaceDrunk替换
			if OutsidersPool[randIdx] == Drunk {
				for {
					if replaceDrunk != "" {
						break
					}
					repeatFlag = false
					randIdxTownsfolk := rand.Intn(len(TownsfolkPool))
					for _, num := range townsfolkRandNums {
						if randIdxTownsfolk == num {
							repeatFlag = true
							break
						}
					}
					if !repeatFlag {
						replaceDrunk = TownsfolkPool[randIdxTownsfolk]
						break
					}
				}
			}
		}
	}
	for _, randIdx := range townsfolkRandNums {
		characterPoolForSelection = append(characterPoolForSelection, TownsfolkPool[randIdx])
	}

	// 打乱顺序
	characterPoolForSelection = Shuffle(characterPoolForSelection)

	var characterTypePoolForSelection []string
	for _, elem := range characterPoolForSelection {
		for key, pool := range CharacterPool {
			if Contains(pool, elem) {
				characterTypePoolForSelection = append(characterTypePoolForSelection, key)
			}
		}
	}

	var newPlayers []model.Player
	var newPlayer model.Player
	for i := range players {
		newPlayer.Id = players[i].Id
		newPlayer.Name = players[i].Name
		newPlayer.Index = i
		newPlayer.Waiting = false
		newPlayer.Character = characterPoolForSelection[i]
		newPlayer.CharacterType = characterTypePoolForSelection[i]
		newPlayers = append(newPlayers, newPlayer)
	}

	return newPlayers, replaceDrunk
}

func genRandomPositionSlice(indexSliceForCharacterTypePool []int, characterByTypePool []string, num int) []int {
	randomInt := rand.Intn(len(characterByTypePool))
	indexSliceForCharacterTypePool = append(indexSliceForCharacterTypePool, randomInt)
	for {
		if len(indexSliceForCharacterTypePool) == num {
			break
		}
		randomInt = rand.Intn(len(characterByTypePool))
		repeatFlag := false
		for j := 0; j < len(indexSliceForCharacterTypePool); j++ {
			if indexSliceForCharacterTypePool[j] == randomInt {
				repeatFlag = true
				break
			}
		}
		if !repeatFlag {
			indexSliceForCharacterTypePool = append(indexSliceForCharacterTypePool, randomInt)
		}
	}
	return indexSliceForCharacterTypePool
}

func initStatus(players []model.Player, replaceDrunk string) []model.Player {
	for i, player := range players {
		players[i].Ready.Nominate = true
		players[i].Ready.Nominated = true
		players[i].Ready.Vote = 1
		switch player.Character {
		case Imp:
			players[i].State.Evil = true
			players[i].State.Demon = true
		case Poisoner:
			players[i].State.Evil = true
		case ScarletWoman:
			players[i].State.Evil = true
		case Baron:
			players[i].State.Evil = true
		case Virgin:
			players[i].State.Blessed = true
		case Slayer:
			players[i].State.Bullet = true
		case Recluse:
			players[i].State.Evil = true
			evilCharacter := getRandEvilCharacter()
			players[i].State.RegardedAs = evilCharacter
			if evilCharacter == Spy {
				players[i].State.Evil = false
			}
			players[i].State.RegardedAsSaved = evilCharacter
			if players[i].State.RegardedAsSaved == Imp {
				players[i].State.Demon = true
			}
		case Drunk:
			players[i].CharacterType = Townsfolk
			players[i].Character = replaceDrunk
			players[i].State.Drunk = true
		case FortuneTeller:
			for {
				randIdx := rand.Intn(len(players))
				if players[randIdx].CharacterType == Townsfolk {
					players[randIdx].State.Demon = true
					break
				}
			}
		}
	}

	return players
}

func getRandEvilCharacter() string {
	evils := append(MinionsPool, DemonsPool...)
	randInt := rand.Intn(len(evils))
	return evils[randInt]
}

func quitGame(mux *sync.RWMutex, room *model.Room, playerId string) {
	cfg := model.GetConfig()
	mux.Lock()
	defer mux.Unlock()
	if room == nil {
		return
	}

	for i, player := range room.Players {
		if player.Id == playerId {
			room.Players[i].Quited = true
			break
		}
	}

	// 关闭退出者的gaming连接
	if cfg.GamingConnPool[room.Id] == nil {
		return // 防空指针异常
	}
	for id, conn := range cfg.GamingConnPool[room.Id] {
		if id == playerId {
			CfgMutex.Lock()
			conn.Close()
			delete(cfg.GamingConnPool[room.Id], id)
			CfgMutex.Unlock()
			continue
		}
	}
	// 关闭退出者的game连接
	if cfg.GameConnPool[room.Id] == nil {
		return // 防空指针异常
	}
	for id, conn := range cfg.GameConnPool[room.Id] {
		if id == playerId {
			CfgMutex.Lock()
			conn.Close()
			delete(cfg.GameConnPool[room.Id], id)
			CfgMutex.Unlock()
			continue
		}
	}
}

func detectIfAllQuited(mux *sync.RWMutex, room *model.Room) {
	cfg := model.GetConfig()
	mux.Lock()
	defer mux.Unlock()

	if room == nil {
		return
	}

	var allQuited = true
	for _, player := range room.Players {
		allQuited = allQuited && player.Quited
	}
	if allQuited {
		CfgMutex.Lock()
		var newRooms []model.Room
		for _, roomm := range cfg.Rooms {
			if room.Id != roomm.Id {
				newRooms = append(newRooms, roomm)
			}
		}
		cfg.Rooms = newRooms
		CfgMutex.Unlock()
	}
}
