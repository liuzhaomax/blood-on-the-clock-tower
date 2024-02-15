package model

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
