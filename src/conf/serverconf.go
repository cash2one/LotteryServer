package conf

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

var ServerConfig struct {
	LogLevel      string //Log等级
	LogPath       string //Log路径
	LocalHttpURL  string //本地Http URL
	RemoteHttpURL string //远程连接的Http URL

	UserName string //管理员用户名
	Password string //管理员密码

	ReturnRate         int //返奖率
	ExpectedRate       int //期望中奖率
	SingleBettingScore int //单注积分
	MaxBettingNum	   int //最大投注积分
	CurrentWinRate     int //实际中奖率
	RewardNum0         int //奖励0数量
	RewardScore0       int //奖励0分数
	RewardNum1         int //奖励1数量
	RewardScore1       int //奖励1分数
	RewardNum2         int //奖励2数量
	RewardScore2       int //奖励2分数
	RewardNum3         int //奖励3数量
	RewardScore3       int //奖励3分数
	RewardNum4         int //奖励4数量
	RewardScore4       int //奖励4分数
	RewardNum5         int //奖励5数量
	RewardScore5       int //奖励5分数
	RewardNum6		   int //奖励6数量
	RewardScore6	   int //奖励6分数

	DBHost string	//数据库IP
	DBUser string	//数据库用户
	DBPasswd string	//数据库密码
}

func init() {
	data, err := ioutil.ReadFile("conf/server.json")
	if err != nil {
		log.Fatal("%v", err)
		return
	}

	err = json.Unmarshal(data, &ServerConfig)
	if err != nil {
		log.Fatal("%v", err)
		return
	}
}

func UpdateConfig(){
	data, err := json.Marshal(&ServerConfig)
	if err != nil{
		log.Fatal("%v", err)
		return
	}

	ioutil.WriteFile("conf/server.json", data, 0666)
}
