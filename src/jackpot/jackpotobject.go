package jackpot

import (
	"conf"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

//获取单例
var _instance *JackpotObject

func JackpotInstance() *JackpotObject {
	if _instance == nil {
		_instance = NewJackpotObject()
		_instance.GenerateJackpot()
	}

	return _instance
}

func NewJackpotObject() *JackpotObject {
	myJackpot := &JackpotObject{RewardDataInfo: make([]BettingReward, 7)}

	//初始化基本参数
	myJackpot.ReturnRate = 9000
	myJackpot.ExpectedRate = 3500
	myJackpot.SingleBettingScore = 10
	myJackpot.MaxBettingNum = 10
	myJackpot.overflowNum = 0
	myJackpot.CurrentWinRate = 0

	//初始化奖池
	myJackpot.RewardDataInfo[0].RewardNum = 0
	myJackpot.RewardDataInfo[0].RewardScore = 0
	myJackpot.RewardDataInfo[1].RewardNum = 1
	myJackpot.RewardDataInfo[1].RewardScore = 1 * 10
	myJackpot.RewardDataInfo[2].RewardNum = 500
	myJackpot.RewardDataInfo[2].RewardScore = 2 * 10
	myJackpot.RewardDataInfo[3].RewardNum = 400
	myJackpot.RewardDataInfo[3].RewardScore = 3 * 10
	myJackpot.RewardDataInfo[4].RewardNum = 300
	myJackpot.RewardDataInfo[4].RewardScore = 5 * 10
	myJackpot.RewardDataInfo[5].RewardNum = 200
	myJackpot.RewardDataInfo[5].RewardScore = 10 * 10
	myJackpot.RewardDataInfo[6].RewardNum = 100
	myJackpot.RewardDataInfo[6].RewardScore = 20 * 10

	return myJackpot
}

type BettingReward struct {
	RewardScore int //单注奖项的分数
	RewardNum   int //奖项的数量
}

type JackpotObject struct {
	lock               sync.Mutex      //奖池的锁，所有操作都需要加锁
	ReturnRate         int             //返奖率
	ExpectedRate       int             //期望中奖率
	CurrentWinRate     int             //奖池实际中奖率
	SingleBettingScore int             //单注积分
	MaxBettingNum      int             //最高下注数量
	overflowNum        int             //彩票溢出数量
	remainNum          int             //剩余的彩票数量
	RewardDataInfo     []BettingReward //奖项的信息
}

//重新生成奖池
func (this *JackpotObject) GenerateJackpot() {
	//先加锁
	this.lock.Lock()
	defer this.lock.Unlock()

	this.regenerate_jackpot()
}

//玩家抽奖,返回中奖的分数
func (this *JackpotObject) LotteryDraw(lotteryMultiple, lotteryType int) int {
	//先加锁
	this.lock.Lock()
	defer this.lock.Unlock()

	stRewardData := this.RewardDataInfo

	var iRewardScore int = 0
	var bWinReward bool = false
	var iGroupNum int = 0
	var iRandResult int = 0

	//如果已经抽完，重新生成奖池
	if this.remainNum <= 0 {
		this.regenerate_jackpot()
	}

	//重新设置随机数种子
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	//lotteryType := r.Intn(7)
	if lotteryType == 2 {
		//类型为2需要抽奖 0-3需要抽奖的情况
		iGroupNum = (stRewardData[0].RewardNum + stRewardData[1].RewardNum + stRewardData[2].RewardNum + stRewardData[3].RewardNum) / lotteryMultiple
		iRandResult = iGroupNum*r.Intn(100)/100 + 1
		if iRandResult > (stRewardData[0].RewardNum / lotteryMultiple) {
			bWinReward = true
			iRandResult = iRandResult - (stRewardData[0].RewardNum / lotteryMultiple)
			if iRandResult <= (stRewardData[3].RewardNum / lotteryMultiple) {
				stRewardData[3].RewardNum -= lotteryMultiple
				iRewardScore = lotteryMultiple * stRewardData[3].RewardScore
			} else {
				iRandResult = iRandResult - (stRewardData[3].RewardNum / lotteryMultiple)
				if iRandResult <= (stRewardData[2].RewardNum / lotteryMultiple) {
					stRewardData[2].RewardNum -= lotteryMultiple
					iRewardScore = lotteryMultiple * stRewardData[2].RewardScore
				} else {
					stRewardData[1].RewardNum -= lotteryMultiple
					iRewardScore = lotteryMultiple * stRewardData[1].RewardScore
				}
			}
		}
	} else if lotteryType == 4 {
		//悟空
		if lotteryMultiple > stRewardData[6].RewardNum {
			if this.overflowNum > 0 {
				lotteryType = 6
				NumVacancy := lotteryMultiple - stRewardData[lotteryType].RewardNum
				ScoreVacancy := NumVacancy * stRewardData[lotteryType].RewardScore
				numNeeded := ScoreVacancy / stRewardData[1].RewardScore
				if this.overflowNum >= numNeeded && stRewardData[1].RewardNum >= numNeeded {
					stRewardData[1].RewardNum -= numNeeded
					this.overflowNum -= (numNeeded - lotteryMultiple)
					bWinReward = true
					iRewardScore = lotteryMultiple * stRewardData[lotteryType].RewardScore
				} else {
					numNeeded := ScoreVacancy / stRewardData[2].RewardScore
					if this.overflowNum >= numNeeded && stRewardData[2].RewardNum >= numNeeded {
						stRewardData[2].RewardNum -= numNeeded
						this.overflowNum -= (numNeeded - lotteryMultiple)
						bWinReward = true
						iRewardScore = lotteryMultiple * stRewardData[lotteryType].RewardScore
					} else {
						numNeeded := ScoreVacancy / stRewardData[3].RewardScore
						if this.overflowNum >= numNeeded && stRewardData[3].RewardNum >= numNeeded {
							stRewardData[3].RewardNum -= numNeeded
							this.overflowNum -= (numNeeded - lotteryMultiple)
							bWinReward = true
							iRewardScore = lotteryMultiple * stRewardData[lotteryType].RewardScore
						}
					}
				}
			}
		}

		if !bWinReward {
			//没中奖
			if lotteryMultiple > stRewardData[5].RewardNum {
				if stRewardData[0].RewardNum <= 0 {
					lotteryType = 5
					NumVacancy := lotteryMultiple - stRewardData[lotteryType].RewardNum
					ScoreVacancy := NumVacancy * stRewardData[lotteryType].RewardScore
					numNeeded := ScoreVacancy / stRewardData[1].RewardScore
					if this.overflowNum >= numNeeded && stRewardData[1].RewardNum >= numNeeded {
						stRewardData[1].RewardNum -= numNeeded
						this.overflowNum -= (numNeeded - lotteryMultiple)
						bWinReward = true
						iRewardScore = lotteryMultiple * stRewardData[lotteryType].RewardScore
					}
				}
			}
		}

		if !bWinReward {
			//没中奖
			iGroupNum = (stRewardData[0].RewardNum + stRewardData[5].RewardNum + stRewardData[6].RewardNum) / lotteryMultiple
			iRandResult = iGroupNum*rand.Intn(100)/100 + 1
			if iRandResult > (stRewardData[0].RewardNum / lotteryMultiple) {
				iRandResult = iRandResult - (stRewardData[0].RewardNum / lotteryMultiple)
				if iRandResult <= (stRewardData[5].RewardNum / lotteryMultiple) {
					bWinReward = true
					stRewardData[5].RewardNum -= lotteryMultiple
					iRewardScore = lotteryMultiple * stRewardData[5].RewardScore
				} else {
					iRandResult = iRandResult - (stRewardData[5].RewardNum / lotteryMultiple)
					if iRandResult <= (stRewardData[6].RewardNum / lotteryMultiple) {
						bWinReward = true
						stRewardData[6].RewardNum -= lotteryMultiple
						iRewardScore = lotteryMultiple * stRewardData[6].RewardScore
					}
				}
			}
		}
	} else if lotteryType == 3 || lotteryType == 5 {
		//八戒或者沙僧
		if lotteryMultiple > stRewardData[5].RewardNum {
			if this.overflowNum > 0 {
				lotteryType = 5
				NumVacancy := lotteryMultiple - stRewardData[lotteryType].RewardNum
				ScoreVacancy := NumVacancy * stRewardData[lotteryType].RewardScore
				numNeeded := ScoreVacancy / stRewardData[1].RewardScore
				if this.overflowNum >= numNeeded && stRewardData[1].RewardNum >= numNeeded {
					stRewardData[1].RewardNum -= numNeeded
					this.overflowNum -= (numNeeded - lotteryMultiple)
					bWinReward = true
					iRewardScore = lotteryMultiple * stRewardData[lotteryType].RewardScore
				} else {
					numNeeded := ScoreVacancy / stRewardData[2].RewardScore
					if this.overflowNum >= numNeeded && stRewardData[2].RewardNum >= numNeeded {
						stRewardData[2].RewardNum -= numNeeded
						this.overflowNum -= (numNeeded - lotteryMultiple)
						bWinReward = true
						iRewardScore = lotteryMultiple * stRewardData[lotteryType].RewardScore
					} else {
						numNeeded := ScoreVacancy / stRewardData[3].RewardScore
						if this.overflowNum >= numNeeded && stRewardData[3].RewardNum >= numNeeded {
							stRewardData[3].RewardNum -= numNeeded
							this.overflowNum -= (numNeeded - lotteryMultiple)
							bWinReward = true
							iRewardScore = lotteryMultiple * stRewardData[lotteryType].RewardScore
						}
					}
				}
			}
		}

		if !bWinReward {
			//没中奖
			if lotteryMultiple > stRewardData[4].RewardNum {
				if stRewardData[0].RewardNum <= 0 {
					lotteryType = 4
					NumVacancy := lotteryMultiple - stRewardData[lotteryType].RewardNum
					ScoreVacancy := NumVacancy * stRewardData[lotteryType].RewardScore
					numNeeded := ScoreVacancy / stRewardData[1].RewardScore
					if this.overflowNum >= numNeeded && stRewardData[1].RewardNum >= numNeeded {
						stRewardData[1].RewardNum -= numNeeded
						this.overflowNum -= (numNeeded - lotteryMultiple)
						bWinReward = true
						iRewardScore = lotteryMultiple * stRewardData[lotteryType].RewardScore
					}
				}
			}
		}

		if !bWinReward {
			//没中奖
			iGroupNum = (stRewardData[0].RewardNum + stRewardData[4].RewardNum + stRewardData[5].RewardNum) / lotteryMultiple
			iRandResult = iGroupNum*rand.Intn(100)/100 + 1
			if iRandResult > (stRewardData[0].RewardNum / lotteryMultiple) {
				iRandResult = iRandResult - (stRewardData[0].RewardNum / lotteryMultiple)
				if iRandResult <= (stRewardData[4].RewardNum / lotteryMultiple) {
					bWinReward = true
					stRewardData[4].RewardNum -= lotteryMultiple
					iRewardScore = lotteryMultiple * stRewardData[4].RewardScore
				} else {
					iRandResult = iRandResult - (stRewardData[4].RewardNum / lotteryMultiple)
					if iRandResult <= (stRewardData[5].RewardNum / lotteryMultiple) {
						bWinReward = true
						stRewardData[5].RewardNum -= lotteryMultiple
						iRewardScore = lotteryMultiple * stRewardData[5].RewardScore
					}
				}
			}
		}
	}

	if !bWinReward {
		stRewardData[0].RewardNum -= lotteryMultiple
		if stRewardData[0].RewardNum < 0 {
			this.overflowNum -= stRewardData[0].RewardNum
			stRewardData[0].RewardNum = 0
		}
	}

	this.remainNum -= lotteryMultiple

	return iRewardScore
}

//生成奖池
func (this *JackpotObject) regenerate_jackpot() {
	fmt.Printf("jackpot info: %v\n", this)

	//化基本参数
	this.ReturnRate = conf.ServerConfig.ReturnRate
	this.ExpectedRate = conf.ServerConfig.ExpectedRate
	this.SingleBettingScore = conf.ServerConfig.SingleBettingScore
	this.MaxBettingNum = conf.ServerConfig.MaxBettingNum
	this.overflowNum = 0
	this.CurrentWinRate = 0

	//初始化奖池
	this.RewardDataInfo[0].RewardNum = 0
	this.RewardDataInfo[0].RewardScore = conf.ServerConfig.RewardScore0
	this.RewardDataInfo[1].RewardNum = 1
	this.RewardDataInfo[1].RewardScore = conf.ServerConfig.RewardScore1
	this.RewardDataInfo[2].RewardNum = conf.ServerConfig.RewardNum2
	this.RewardDataInfo[2].RewardScore = conf.ServerConfig.RewardScore2
	this.RewardDataInfo[3].RewardNum = conf.ServerConfig.RewardNum3
	this.RewardDataInfo[3].RewardScore = conf.ServerConfig.RewardScore3
	this.RewardDataInfo[4].RewardNum = conf.ServerConfig.RewardNum4
	this.RewardDataInfo[4].RewardScore = conf.ServerConfig.RewardScore4
	this.RewardDataInfo[5].RewardNum = conf.ServerConfig.RewardNum5
	this.RewardDataInfo[5].RewardScore = conf.ServerConfig.RewardScore5
	this.RewardDataInfo[6].RewardNum = conf.ServerConfig.RewardNum6
	this.RewardDataInfo[6].RewardScore = conf.ServerConfig.RewardScore6

	stRewardData := this.RewardDataInfo

	//奖池的总金额和总票数
	var iTotalRewardScore, iTotalRewardNum int = 0, 0
	for i := 2; i < 7; i++ {
		iTotalRewardScore += (stRewardData[i].RewardScore * stRewardData[i].RewardNum)
		iTotalRewardNum += stRewardData[i].RewardNum
	}

	var hasResult bool = false
	var currentStep int = 0

	var CurrentWinRate int = 0
	var currentTotalScore, currentTotalNum int = 0, 0
	var iTotalCycleNum int = 0
	for !hasResult {
		iTotalCycleNum++
		if iTotalCycleNum >= 100000 {
			//超过10万次，直接跳出
			break
		}

		//hasResult = true
		currentTotalScore = iTotalRewardScore + (stRewardData[1].RewardScore * stRewardData[1].RewardNum)
		//fmt.Println("current total score: ", currentTotalScore)
		currentTotalNum = currentTotalScore * 10000 / this.ReturnRate / this.SingleBettingScore
		//fmt.Printf("current total num:%d, return rate:%d, single betting score:%d\n", currentTotalNum, this.returnRate, this.singleBettingScore)
		CurrentWinRate = (iTotalRewardNum + stRewardData[1].RewardNum) * 10000 / currentTotalNum
		//fmt.Printf("current win rate:%d, total reward num:%d,current total num %d\n", CurrentWinRate, iTotalRewardNum, currentTotalNum)
		if CurrentWinRate == this.ExpectedRate {
			hasResult = true
		} else if CurrentWinRate < this.ExpectedRate {
			if currentStep == 0 {
				stRewardData[1].RewardNum *= 2
				stRewardData[1].RewardScore = 10
			} else {
				hasResult = true
			}
		} else {
			currentStep = 1
			stRewardData[1].RewardNum -= 1
			stRewardData[1].RewardScore = 10
		}
	}

	this.CurrentWinRate = CurrentWinRate
	stRewardData[0].RewardNum = currentTotalNum - (iTotalRewardNum + stRewardData[1].RewardNum)

	this.remainNum = 0
	for i := 0; i < 7; i++ {
		this.remainNum += stRewardData[i].RewardNum
	}

	//fmt.Printf("jackpot info: %v\n", this)

	return
}

func init() {
	JackpotInstance()
}
