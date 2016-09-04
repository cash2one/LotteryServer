package msghelper

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"crypto/rand"
	"math/big"

	"github.com/bitly/go-simplejson"
)

const (
	HTTP_URL         = "http://180.168.15.19:20081/api"     //Http Post的URL
	LOTTERY_HTTP_URL = "http://127.0.0.1:8080/lotteryapi" //抽奖的Http Post URL
	GET_TOKEN        = "1001"                                //获取TOKEN
	GET_ACCOUNT_INFO = "3001"                                //账户信息查询
	DO_LOTTERY_DRAW  = "7001"                                //玩家抽奖
	ADD_PLAYER_SCORE = "6001"                                //玩家积分操作
)

type ClientRequest struct{
	Token 	string
	NickId	string
	ChipId	string
	ValidateData string
	Machineid string
	AgentId string
	OrderId string
}

type MsgHandler struct {
	strToken string
}

//获取用户token
func (this *MsgHandler) GetUserToken() string {
	js, _ := simplejson.NewJson([]byte(`{}`))

	//设置请求内容
	js.SetPath([]string{"header", "cmd"}, DO_LOTTERY_DRAW)

	js.SetPath([]string{"body", "token"}, "4rminqlhl9uiucffk92grcks")
	js.SetPath([]string{"body", "userID"}, "0200002400008")
	js.SetPath([]string{"body", "cardID"}, "02000024")
	js.SetPath([]string{"body", "machineID"}, "020002")
	js.SetPath([]string{"body", "deviceID"}, "020002")
	js.SetPath([]string{"body", "lotteryMultiple"}, 1)
	js.SetPath([]string{"body", "lotteryType"}, 0)

	userTokenMsg, _ := js.Encode()
	body := this.sendHttpRequest(LOTTERY_HTTP_URL, string(userTokenMsg))	
	//this.SendRequestMsg(GET_TOKEN, "860727010321682", string(userTokenMsg), true))
	//if err != nil {
	//	return ""
	//}

	bodyjs, _ := simplejson.NewJson([]byte(body))
	strToken, _ := bodyjs.String()

	//return this.strToken
	return strToken
}

//拉取账户信息
func (this *MsgHandler) GetAccountInfo(strCardID string) string {
	js, _ := simplejson.NewJson([]byte(`{}`))

	//设置请求内容
	js.Set("token", this.strToken)
	js.Set("cardId", strCardID)

	//发送请求
	getaccountInfoMsg, _ := js.Encode()
	body, err := this.ParseResponseMsg(this.SendRequestMsg(GET_ACCOUNT_INFO, "860727010321682", string(getaccountInfoMsg), false))
	if err != nil {
		return ""
	}

	//todo jasonxiong 这边后面添加实际的逻辑处理
	//bodyjs, _ := simplejson.NewJson([]byte(body))
	//this.strToken, _ = bodyjs.Get("token").String()

	return body
}

func (this *MsgHandler) DoLotteryDraw(strCardID string, iMultiple, iType int) string {
	js, _ := simplejson.NewJson([]byte(`{}`))

	//设置请求的内容
	js.Set("token", this.strToken)
	js.Set("cardId", strCardID)
	js.Set("multiple", iMultiple)
	js.Set("type", iType)

	//发送请求
	lotteryDrawMsg, _ := js.Encode()
	body, err := this.ParseResponseMsg(this.SendRequestMsg(DO_LOTTERY_DRAW, "860727010321682", string(lotteryDrawMsg), true))
	if err != nil {
		return ""
	}

	//打印body的内容
	fmt.Println("lottery draw response body: ", body)

	return body
}

//发送请求数据
func (this *MsgHandler) SendRequestMsg(strCmd, strDeviceID, strRequest string, isLotteryRequest bool) string {
	js, _ := simplejson.NewJson([]byte(`{}`))

	//设置header
	js.SetPath([]string{"header", "messageId"}, this.getNowTime())
	js.SetPath([]string{"header", "cmd"}, strCmd)
	js.SetPath([]string{"header", "deviceId"}, strDeviceID)
	js.SetPath([]string{"header", "sign"}, "0ca175b9c0f726a831d895e269332461")

	fmt.Println("Send Request Msg Body: ", strRequest)
	fmt.Println("encode token: ", this.strToken)

	//设置Body
	if strCmd != GET_TOKEN {
		myDes3Helper := &Des3Helper{m_strKey: this.strToken, m_strIV: ConstStrIV}
		strRequest = myDes3Helper.Encode(strRequest)
	}

	js.Set("body", strRequest)

	//发送数据
	sendData, _ := js.Encode()

	if isLotteryRequest {
		return this.sendHttpRequest(LOTTERY_HTTP_URL, string(sendData))
	} else {
		return this.sendHttpRequest(HTTP_URL, string(sendData))
	}

	return ""
}

//分析response消息，返回body消息的json字符串
func (this *MsgHandler) ParseResponseMsg(strResponseMsg string) (string, error) {
	respjs, err := simplejson.NewJson([]byte(strResponseMsg))
	if err != nil {
		fmt.Errorf("Failed to parse response msg!")
		return "", err
	}

	//获取返回状态
	status, err := respjs.GetPath("header", "resultStatus").String()
	if err != nil || status != "0100" {
		fmt.Errorf("Failed to process msg, invlaid return status")
		return "", errors.New("Invalid resultStatus")
	}

	//获取CMD
	cmd, err := respjs.GetPath("header", "cmd").String()
	body, _ := respjs.Get("body").String()
	if cmd == GET_TOKEN {
		//拉取用户token的消息
		return body, nil
	} else {
		//其他正常消息
		myDes3Helper := &Des3Helper{m_strKey: this.strToken, m_strIV: ConstStrIV}
		return myDes3Helper.Decode(body), nil
	}

	return "", errors.New("Invalid Response")
}

//分析request消息，返回cmd 和 body消息的json字符串
func (this *MsgHandler) ParseRequestMsg(strRequestMsg string) (string, string, error) {
	reqjs, err := simplejson.NewJson([]byte(strRequestMsg))
	if err != nil {
		fmt.Errorf("Failed to parse request msg!")
		return "", "", err
	}

	strBody, err := reqjs.Get("body").String()
	if err != nil {
		fmt.Errorf("Failed to parse body msg")
		return "", "", err
	}

	cmd, _ := reqjs.GetPath("header", "cmd").String()
	if cmd != GET_TOKEN {
		//body DES3解码
		myDes3Helper := &Des3Helper{m_strKey: this.strToken, m_strIV: ConstStrIV}
		strBody = myDes3Helper.Decode(strBody)
	}

	return cmd, strBody, nil
}

func (this *MsgHandler) AddAccountScore(newRequest ClientRequest, deviceID string, addScore int) error {
	if addScore == 0 {
		return nil
	}

	js, _ := simplejson.NewJson([]byte(`{}`))

	//设置请求内容
	js.Set("token", newRequest.Token)
	js.Set("nickId", newRequest.NickId)
	js.Set("chipId", newRequest.ChipId)
	js.Set("validateData", newRequest.ValidateData)
	js.Set("machineid", newRequest.Machineid)
	js.Set("agentId", newRequest.AgentId)
	js.Set("orderId", newRequest.OrderId)
	js.Set("gameId", "001")

	if addScore >= 0 {
		//增加积分
		js.Set("getType", 1)
		js.Set("integralGetGoal", addScore)
	} else {
		//扣除积分
		js.Set("getType", 2)
		js.Set("integralGetGoal", -addScore)
	}

	//发送请求
	strAddScoreMsg, _ := js.Encode()
	fmt.Println("str AddScore Msg:  ", string(strAddScoreMsg))
	_, err := this.ParseResponseMsg(this.SendRequestMsg(ADD_PLAYER_SCORE, deviceID, string(strAddScoreMsg), false))
	if err != nil {
		fmt.Println("Failed to add user score!")
		return err
	}else{
		fmt.Println("Success to add user score!")
		return nil
	}

	return nil
}

func (this *MsgHandler) EncodeLotteryResponse(rewardScore int) string {
	js, _ := simplejson.NewJson([]byte(`{}`))

	//设置内容
	js.Set("reward", rewardScore)

	strRespBody, _ := js.Encode()
	//myDes3Helper := &Des3Helper{m_strKey: this.strToken, m_strIV: ConstStrIV}

	//return myDes3Helper.Encode(string(strRespBody))
	return string(strRespBody)
}

func (this *MsgHandler) SetToken(token_str string) {
	this.strToken = token_str
}

//获取当前时间
func (this *MsgHandler) getNowTime() string {
	rnd,_ := rand.Int(rand.Reader,big.NewInt(1000000))
	return time.Now().Format("20060102150405") + fmt.Sprintf("%d",rnd)
}

//发送http请求
func (this *MsgHandler) sendHttpRequest(strURL, strPostData string) string {
	client := &http.Client{}
	req, err := http.NewRequest("POST", strURL, strings.NewReader(strPostData))
	if err != nil {
		fmt.Errorf("Failed to new http request")
		return ""
	}

	req.Header.Set("Content-Type", "application/json")

	fmt.Println("http url: ", strURL)
	fmt.Println("post data: ", strPostData)

	resp, err := client.Do(req)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Errorf("Failed to get http response msg")
		return ""
	}

	fmt.Println("resp msg:", string(body))

	return string(body)
}
