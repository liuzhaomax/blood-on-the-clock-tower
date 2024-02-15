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
	Status
}

type Status struct {
	Dead  bool `json:"dead"`
	Evil  bool `json:"evil"`
	Demon bool `json:"demon"`
}
