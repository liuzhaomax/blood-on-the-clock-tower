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
		cfg.HomeConnPool = map[string]*websocket.Conn{}
		cfg.RoomConnPool = map[string]map[string]*websocket.Conn{}
		cfg.GameConnPool = map[string]map[string]*websocket.Conn{}
	})
}

func GetConfig() *Config {
	return cfg
}

type Config struct {
	Rooms        []Room
	HomeConnPool map[string]*websocket.Conn            // 首页 玩家长连接 [playId]conn
	RoomConnPool map[string]map[string]*websocket.Conn // 等待开始 玩家的长连接 [roomId][playId]conn
	GameConnPool map[string]map[string]*websocket.Conn // 游戏中玩家的game长连接 [roomId][playId]conn
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
	State     GameState `json:"state"` // 游戏中各身份状态
	// 技能施放池，存储所有施放技能人，当前阶段施放的技能作用目标
	CastPool map[string][]string `json:"castPool"` // 本轮施法池 [playId][]targetId{}
	VotePool map[string]string   `json:"votePool"` // 本轮票池 [playId]log{}
	Executed *Player             `json:"executed"` // 本轮被处决者
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
	Waiting       bool        `json:"waiting"` // 是否等待开始
	Quited        bool        `json:"quited"`  // 是否已经退出
	Character     string      `json:"character"`
	CharacterType string      `json:"characterType"`
	State         PlayerState `json:"state"`
	Ready         PlayerReady `json:"ready"`
	Log           string      `json:"log"`
}

type PlayerState struct {
	Dead            bool   `json:"dead"`            // 全体
	Evil            bool   `json:"evil"`            // 隐士
	RegardedAsSaved string `json:"regardedAsSaved"` // 隐士的被视作的身份，记录的
	RegardedAs      string `json:"regardedAs"`      // 隐士的被视作的身份
	Demon           bool   `json:"demon"`           // 占卜
	Drunk           bool   `json:"drunk"`           // 酒鬼
	Poisoned        bool   `json:"poisoned"`        // 下毒
	Protected       bool   `json:"protected"`       // 僧侣
	Bullet          bool   `json:"bullet"`          // 杀手
	Blessed         bool   `json:"blessed"`         // 圣女
	Master          bool   `json:"master"`          // 管家
	Casted          bool   `json:"casted"`          // 本轮是否已施放主动技能
	Nominated       bool   `json:"nominated"`       // 本轮是否已被提名
	Voted           bool   `json:"voted"`           // 本轮是否已投票
	VoteCount       int    `json:"voteCount"`       // 得票数
}

type PlayerReady struct {
	Nominate  bool `json:"nominate"`  // 可以提名
	Nominated bool `json:"nominated"` // 可以被提名
	Vote      int  `json:"vote"`      // 可以投票的票数
}
