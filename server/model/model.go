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
	ConnPool map[string]*websocket.Conn
}

type Room struct {
	Id       string    `json:"id"`
	Name     string    `json:"name"`
	Password string    `json:"password"`
	Host     string    `json:"host"`
	Status   string    `json:"status"`
	Init     bool      `json:"init"`
	Players  []Player  `json:"players"`
	Log      string    `json:"log"`
	State    GameState `json:"state"`
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
}

type PlayerReady struct {
	// 不同于playerState中的技能，这个是每个阶段的是否放技能的情况，
	// 也就是说只要要进入的阶段是不能发动技能的，那这个属性直接被赋为true，
	// 下个阶段开始会变为false
	Casted    bool `json:"casted"`    // 已施法
	Nominate  bool `json:"nominate"`  // 可以提名
	Nominated bool `json:"nominated"` // 可以被提名
	Vote      int  `json:"vote"`      // 可以投票
	VoteCount int  `json:"voteCount"` // 得票数
}

type ActionReq struct {
	Action  string   `json:"action"`
	Targets []string `json:"targets"`
}
