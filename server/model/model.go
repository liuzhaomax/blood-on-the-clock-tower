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
	Id       string   `json:"id"`
	Name     string   `json:"name"`
	Password string   `json:"password"`
	Host     string   `json:"host"`
	State    string   `json:"state"`
	Init     bool     `json:"init"`
	Night    bool     `json:"night"`
	Day      int      `json:"day"`
	Players  []Player `json:"players"`
	Log      string   `json:"log"`
}

type Player struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Index         int    `json:"index"`
	Character     string `json:"character"`
	CharacterType string `json:"characterType"`
	Status        Status `json:"status"`
	Log           string `json:"log"`
}

type Status struct {
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
