package jackpot

import (
	"testing"
	"fmt"
)

func Test_NewJackpot(t *testing.T) {
	
	//不测试生成奖池
	return

	myJackpot := NewJackpotObject()
	myJackpot.ExpectedRate = 3500
	myJackpot.ReturnRate = 9000
	myJackpot.SingleBettingScore = 10
	myJackpot.MaxBettingNum = 10
	myJackpot.RewardDataInfo[2].RewardNum = 800
	myJackpot.RewardDataInfo[2].RewardScore = 2 * 10
	myJackpot.RewardDataInfo[3].RewardNum = 400
	myJackpot.RewardDataInfo[3].RewardScore = 3 * 10
	myJackpot.RewardDataInfo[4].RewardNum = 200
	myJackpot.RewardDataInfo[4].RewardScore = 10 * 10
	myJackpot.RewardDataInfo[5].RewardNum = 100
	myJackpot.RewardDataInfo[5].RewardScore = 50 * 10

	myJackpot.GenerateJackpot()

	//输出实际中奖率
	t.Log("current win rate: ", myJackpot.CurrentWinRate)

	//输出彩票的注数
	for i := 0; i < 6; i++ {
		t.Logf("rewardNum:rewardscore is %d:%d", myJackpot.RewardDataInfo[i].RewardNum, myJackpot.RewardDataInfo[i].RewardScore)
	}

	//重新生成一次
	myJackpot.GenerateJackpot()
	//输出实际中奖率
	t.Log("current win rate: ", myJackpot.CurrentWinRate)

	//输出彩票的注数
	for i := 0; i < 6; i++ {
		t.Logf("rewardNum:rewardscore is %d:%d", myJackpot.RewardDataInfo[i].RewardNum, myJackpot.RewardDataInfo[i].RewardScore)
	}

	//抽奖
	for i := 0; i < 1000; i++ {
		iRewardScore := myJackpot.LotteryDraw(50, 3)
		t.Log("reward multiple is ", iRewardScore)
	}
}

func Test_DoJackpot(t *testing.T){
	myJackpot := NewJackpotObject()
	myJackpot.ExpectedRate = 2000
	myJackpot.ReturnRate = 9600
	myJackpot.SingleBettingScore = 10
	myJackpot.MaxBettingNum = 10
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

	myJackpot.GenerateJackpot()

	//fmt.Println("After this line")

	i := 0
	rewardNum := 0
	for myJackpot.remainNum >= 0{
		rewardNum = myJackpot.LotteryDraw(1,i%6)
		fmt.Println(rewardNum)

		i = i+1
	}

	return

	//rewardNum := 0
	//for j:=0; j<6; j++{
		//j := 3
		//for i:=0; i<10000; i++{
			//for j:=0; j<6; j++{
				//rewardNum = 
		//		j := 4
		//		rewardNum = myJackpot.LotteryDraw(1, j)
		//		fmt.Println(rewardNum)
			//}
		//}
	//}

	//fmt.Println("Test Ended")
}
