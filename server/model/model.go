package model

import (
	"github.com/gorilla/websocket"
	"sync"
)

var cfg *Config
var once sync.Once

func init() {
	once.Do(func() {
		cfg = &Config{}
		cfg.ConnPool = map[string]*websocket.Conn{}
	})
}

func GetConfig() *Config {
	return cfg
}

type Config struct {
	Rooms    []Room
	ConnPool map[string]*websocket.Conn // 游戏中玩家的长连接
}

type Room struct {
	Id       string    `json:"id"`
	Name     string    `json:"name"`
	Password string    `json:"password"`
	Host     string    `json:"host"`
	Status   string    `json:"status"` // 等待开始，游戏中，复盘中
	Init     bool      `json:"init"`   // 房间是否已初始化身份
	Players  []Player  `json:"players"`
	Log      string    `json:"log"`    // 总日志
	Result   string    `json:"result"` // 游戏结果
	State    GameState `json:"state"`  // 游戏中各身份状态
	// 技能施放池，存储所有施放技能人，当前阶段施放的技能作用目标
	CastPool map[string][]string `json:"castPool"`
}

type GameState struct {
	Night      bool `json:"night"`
	Day        int  `json:"day"`
	Stage      int  `json:"stage"`
	VotingStep bool `json:"votingStep"` // 投票环节
}

type Player struct {
	Id            string      `json:"id"`
	Name          string      `json:"name"`
	Index         int         `json:"index"`
	Character     string      `json:"character"`
	CharacterType string      `json:"characterType"`
	State         PlayerState `json:"state"`
	Log           string      `json:"log"`
	Ready         PlayerReady `json:"ready"`
}

type PlayerState struct {
	Dead      bool `json:"dead"`      // 全体
	Evil      bool `json:"evil"`      // 隐士
	Demon     bool `json:"demon"`     // 占卜
	Drunk     bool `json:"drunk"`     // 酒鬼
	Poisoned  bool `json:"poisoned"`  // 下毒
	Protected bool `json:"protected"` // 僧侣
	Bullet    bool `json:"bullet"`    // 杀手
	Blessed   bool `json:"blessed"`   // 圣女
	Master    bool `json:"master"`    // 管家
	Casted    bool `json:"casted"`    // 本轮是否已施放主动技能
	Nominated bool `json:"nominated"` // 本轮是否已被提名
	Voted     bool `json:"voted"`     // 本轮是否已投票
	VoteCount int  `json:"voteCount"` // 得票数
}

type PlayerReady struct {
	Nominate  bool `json:"nominate"`  // 可以提名
	Nominated bool `json:"nominated"` // 可以被提名
	Vote      int  `json:"vote"`      // 可以投票的票数
}

type ActionReq struct {
	Action  string   `json:"action"`
	Targets []string `json:"targets"`
}
