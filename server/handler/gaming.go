package handler

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/liuzhaomax/blood-on-the-clock-tower/server/model"
	"log"
	"math/rand"
	"net/http"
	"strings"
)

func LoadGame(w http.ResponseWriter, r *http.Request) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	messageType, _, err := conn.ReadMessage()
	if err != nil {
		if websocket.IsCloseError(err, websocket.CloseGoingAway) {
			log.Println("Client disconnected:", err)
			return
		}
		log.Println("Read error:", err)
		return
	}

	// 获取URL参数
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		http.NotFound(w, r)
		return
	}
	roomId := parts[2]

	cfgMutex.Lock()
	defer cfgMutex.Unlock()
	cfg := model.GetConfig()
	_, roomIndex := findRoom(roomId)

	// 分配身份
	if cfg.Rooms[roomIndex].Players[0].Character == "" {
		cfg.Rooms[roomIndex].Players = allocateCharacter(cfg.Rooms[roomIndex].Players)
	}

	marshaledRoom, err := json.Marshal(cfg.Rooms[roomIndex])
	if err != nil {
		log.Println("JSON marshal error:", err)
		return
	}

	if err = conn.WriteMessage(messageType, marshaledRoom); err != nil {
		log.Println("Write error:", err)
		return
	}
}

func allocateCharacter(players []model.Player) []model.Player {
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
	for _, randIdx := range townsfolkRandNums {
		characterPoolForSelection = append(characterPoolForSelection, TownsfolkPool[randIdx])
	}
	if len(outsidersRandNums) != 0 {
		for _, randIdx := range outsidersRandNums {
			characterPoolForSelection = append(characterPoolForSelection, OutsidersPool[randIdx])
		}
	}
	for _, randIdx := range minionsRandNums {
		characterPoolForSelection = append(characterPoolForSelection, MinionsPool[randIdx])
	}
	for _, randIdx := range demonsRandNums {
		characterPoolForSelection = append(characterPoolForSelection, DemonsPool[randIdx])
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
		newPlayer.Character = characterPoolForSelection[i]
		newPlayer.CharacterType = characterTypePoolForSelection[i]
		newPlayers = append(newPlayers, newPlayer)
	}

	return newPlayers
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