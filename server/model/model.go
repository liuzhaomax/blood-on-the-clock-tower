package model

import (
	"sync"
)

var cfg *Config
var once sync.Once

func init() {
	once.Do(func() {
		cfg = &Config{}
		cfg.HomeConnPool = &sync.Map{}
	})
}

func GetConfig() *Config {
	return cfg
}

type Config struct {
	Rooms        []Room
	HomeConnPool *sync.Map // 首页 玩家长连接 [playId]conn
}

type Room struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Password  string    `json:"password"`
	Host      string    `json:"host"`
	CreatedAt string    `json:"createdAt"`
	Status    string    `json:"status"` // 等待开始，游戏中，复盘中
	Init      bool      `json:"init"`   // 房间是否已初始化身份
	Result    string    `json:"result"` // 游戏结果
	Log       string    `json:"log"`    // 总日志
	Players   []Player  `json:"players"`
	State     GameState `json:"state"`     // 游戏中各身份状态
	Executed  *Player   `json:"executed"`  // 本轮被处决者
	Nominated *Player   `json:"nominated"` // 当前投票环节被提名者
	// 技能施放池，存储所有施放技能人，当前阶段施放的技能作用目标
	CastPool map[string][]string `json:"castPool"` // 本轮施法池 [playId][]targetId{}
	VotePool map[string]int      `json:"votePool"` // 本轮票池 [playId]int{}
	VoteLogs map[string]string   `json:"voteLogs"` // 本轮票池日志 [playId]log{}
	// 通信管理
	GameConnPool *sync.Map   `json:"-"` // game长连接[playId]conn
	Mux          *sync.Mutex `json:"-"` // 业务逻辑使用的锁
	ResMux       *sync.Mutex `json:"-"` // 发送响应使用的锁
}

type GameState struct {
	Night      bool `json:"night"`
	Day        int  `json:"day"` // 0代表第一天，1代表第二天，有个加1的关系
	Stage      int  `json:"stage"`
	VotingStep bool `json:"votingStep"` // 投票环节
}

type Player struct {
	Id            string      `json:"id"`
	Name          string      `json:"name"`
	Index         int         `json:"index"`
	Waiting       bool        `json:"waiting"` // 是否等待开始
	Quited        bool        `json:"quited"`  // 是否已经退出
	Character     string      `json:"character"`
	CharacterType string      `json:"characterType"`
	State         PlayerState `json:"state"`
	Ready         PlayerReady `json:"ready"`
	Log           string      `json:"log"`
}

type PlayerState struct {
	Dead             bool   `json:"dead"`             // 全体
	Evil             bool   `json:"evil"`             // 隐士
	RegardedAsSaved  string `json:"regardedAsSaved"`  // 隐士的被视作的身份，记录的
	RegardedAs       string `json:"regardedAs"`       // 隐士的被视作的身份
	RegisteredAsType string `json:"registeredAsType"` // 间谍 - 村民或外来者
	Demon            bool   `json:"demon"`            // 占卜
	Drunk            bool   `json:"drunk"`            // 酒鬼
	Poisoned         bool   `json:"poisoned"`         // 下毒
	Protected        bool   `json:"protected"`        // 僧侣
	Bullet           bool   `json:"bullet"`           // 杀手
	Blessed          bool   `json:"blessed"`          // 圣女
	Master           bool   `json:"master"`           // 管家
	VotedFromMaster  bool   `json:"votedFromMaster"`  // 管家的主人本轮是否投给该玩家
	VotedFromButler  bool   `json:"votedFromButler"`  // 管家本轮是否投给该玩家
	Casted           bool   `json:"casted"`           // 本轮是否已施放主动技能
}

type PlayerReady struct {
	Nominate  bool `json:"nominate"`  // 可以提名
	Nominated bool `json:"nominated"` // 可以被提名
	Vote      bool `json:"vote"`      // 可以投票
}
