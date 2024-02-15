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
	ID            string
	Name          string
	Character     string
	CharacterType string
	Status
}

type Status struct {
	Dead  bool
	Evil  bool
	Demon bool
}
