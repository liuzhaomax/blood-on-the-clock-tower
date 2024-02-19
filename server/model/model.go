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
}

type GameState struct {
	Night        bool `json:"night"`
	Day          int  `json:"day"`
	Stage        int  `json:"stage"`
	CastingStep  bool `json:"castingStep"`
	VotingStep   bool `json:"votingStep"`
	CheckoutStep bool `json:"checkoutStep"`
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
	Dead      bool   `json:"dead"`
	Nominate  bool   `json:"nominate"`
	Nominated bool   `json:"nominated"`
	Vote      bool   `json:"vote"`
	Evil      bool   `json:"evil"`     // 隐士
	Demon     bool   `json:"demon"`    // 占卜
	Drunk     bool   `json:"drunk"`    // 酒鬼
	Poisoned  bool   `json:"poisoned"` // 下毒
	Master    string `json:"master"`   // 管家
	Bullet    bool   `json:"bullet"`   // 杀手
	Blessed   bool   `json:"blessed"`  // 圣女
}

type PlayerReady struct {
	Casted    bool `json:"casted"`
	Nominated bool `json:"nominated"`
	Voted     bool `json:"voted"`
}
