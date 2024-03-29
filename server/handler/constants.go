package handler

const (
	Townsfolk = "村民"
	Outsiders = "外来者"
	Minions   = "爪牙"
	Demons    = "恶魔"

	Washerwoman   = "洗衣妇"
	Librarian     = "图书管理员"
	Investigator  = "调查员"
	Chef          = "厨师"
	Empath        = "共情者"
	FortuneTeller = "占卜师"
	Undertaker    = "掘墓人"
	Monk          = "僧侣"
	Ravenkeeper   = "守鸦人"
	Virgin        = "圣女"
	Slayer        = "杀手"
	Soldier       = "士兵"
	Mayor         = "市长"

	Butler  = "管家"
	Drunk   = "酒鬼"
	Recluse = "隐士"
	Saint   = "圣徒"

	Poisoner     = "下毒者"
	Spy          = "间谍"
	ScarletWoman = "魅魔"
	Baron        = "男爵"

	Imp = "小恶魔"
)

var TownsfolkPool = []string{
	Washerwoman,
	Librarian,
	Investigator,
	Chef,
	Empath,
	FortuneTeller,
	Undertaker,
	Monk,
	Ravenkeeper,
	Virgin,
	Slayer,
	Soldier,
	Mayor,
}

var OutsidersPool = []string{
	Butler,
	Drunk,
	Recluse,
	Saint,
}

var MinionsPool = []string{
	Poisoner,
	Spy,
	ScarletWoman,
	Baron,
}

var DemonsPool = []string{
	Imp,
}

var CharacterPool = map[string][]string{
	Townsfolk: TownsfolkPool,
	Outsiders: OutsidersPool,
	Minions:   MinionsPool,
	Demons:    DemonsPool,
}

var AllChars = append(append(append(TownsfolkPool, OutsidersPool...), MinionsPool...), DemonsPool...)

const (
	RoomWaitingToGo = "等待开始"
	RoomGaming      = "游戏中"
	RoomReviewing   = "复盘中"
)
