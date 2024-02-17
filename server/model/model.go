package model

import "sync"

var cfg *Config
var once sync.Once

func init() {
	once.Do(func() {
		cfg = &Config{}
	})
}

func GetConfig() *Config {
	return cfg
}

type Config struct {
	Rooms []Room
}

type Room struct {
	Id       string   `json:"id"`
	Name     string   `json:"name"`
	Password string   `json:"password"`
	State    string   `json:"state"`
	Players  []Player `json:"players"`
}

type Player struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Character     string `json:"character"`
	CharacterType string `json:"characterType"`
	Status        Status `json:"status"`
}

type Status struct {
	Dead      bool   `json:"dead"`
	Nominate  bool   `json:"nominate"`
	Nominated bool   `json:"nominated"`
	Vote      bool   `json:"vote"`
	Evil      bool   `json:"evil"`     // 隐士
	Demon     bool   `json:"demon"`    // 占卜
	Dunk      bool   `json:"dunk"`     // 酒鬼
	Poisoned  bool   `json:"poisoned"` // 下毒
	Master    string `json:"master"`   // 管家
	Bullet    bool   `json:"bullet"`   // 杀手
	Blessed   bool   `json:"Blessed"`  // 圣女
}
