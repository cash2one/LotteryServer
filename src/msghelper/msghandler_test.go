package msghelper

import (
	"testing"
)

func Test_GetUserToken(t *testing.T) {
	myMsgHandler := &MsgHandler{}

	//测试获取token
	respMsg := myMsgHandler.GetUserToken()
	t.Log("get token resp msg: ", respMsg)

	//测试拉取帐号信息
	//strCardID := "0200001100010"
	//respMsg = myMsgHandler.GetAccountInfo(strCardID)
	//t.Log("get account info resp msg: ", respMsg)

	//测试抽奖
	//iMultiple := 9
	//iType := 0
	//respMsg = myMsgHandler.DoLotteryDraw(strCardID, iMultiple, iType)
	//t.Log("do lottery draw resp msg: ", respMsg)
}
