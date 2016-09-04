package main

import (
	"conf"
	"crypto/rand"
	"encoding/base64"
	"math/big"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"jackpot"
	"msghelper"
	"net/http"
	"strconv"
	"time"
	"database/sql"
	"encoding/json"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	_ "github.com/go-sql-driver/mysql"
)

var (
	cookieStore = sessions.NewCookieStore([]byte("cookie-secret"))
)

func main() {
	
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	
	//抽奖Handler
	http.HandleFunc("/lotteryapi", LotteryHandler)

	//参数配置Handler
	http.HandleFunc("/admin/login", AdminLoginHandler)
	http.HandleFunc("/admin/config", AdminConfigHandler)
	http.HandleFunc("/admin/query", AdminQueryHandler)

	http.ListenAndServe(":8080", context.ClearHandler(http.DefaultServeMux))
}

func LotteryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		sendLotteryResponse(w, "0200", "", "非法的请求类型", "")
		return
	}

	//获取请求消息
	strRequest, _ := ioutil.ReadAll(r.Body)

	//解析请求
	//myMsgHandler := &msghelper.MsgHandler{}
	//myMsgHandler.SetToken("123456789012345678901234")
	//strCmd, strBody, err := myMsgHandler.ParseRequestMsg(string(strRequest))
	//if err != nil {
	//	sendLotteryResponse(w, "0201", "", "")
	//	return
	//}

	fmt.Printf("client request: %s\n", strRequest)

	//目前只处理抽奖请求
	//if strCmd != msghelper.DO_LOTTERY_DRAW {
	//	sendLotteryResponse(w, "0202", strCmd, "")
	//	return
	//}

	//json解码
	bodyjs, _ := simplejson.NewJson([]byte(strRequest))

	//参数解析
	rnd,_ := rand.Int(rand.Reader,big.NewInt(1000000))
	var stNewRequest msghelper.ClientRequest
	stNewRequest.Token,_ = bodyjs.Get("body").Get("token").String()
	stNewRequest.NickId, _ = bodyjs.Get("body").Get("userID").String()
	stNewRequest.ChipId, _ = bodyjs.Get("body").Get("cardID").String()
	stNewRequest.ValidateData = "1"
	stNewRequest.Machineid, _ = bodyjs.Get("body").Get("machineID").String()
	stNewRequest.AgentId = "1"
	stNewRequest.OrderId =  strconv.FormatInt(time.Now().Unix(),10) + fmt.Sprintf("%d", rnd)
	lotteryMultiple, _ := bodyjs.Get("body").Get("lotteryMultiple").Int()
	lotteryType, _ := bodyjs.Get("body").Get("lotteryType").Int()
	deviceID,_ := bodyjs.Get("body").Get("deviceID").String()
	strCmd := msghelper.DO_LOTTERY_DRAW

	//参数检查
	if lotteryType < 0 && lotteryType > 5 {
		sendLotteryResponse(w, "0203", strCmd, "非法的抽奖类型", "")
		return
	}

	myJackpotObj := jackpot.JackpotInstance()
	if lotteryMultiple <= 0 || lotteryMultiple >= myJackpotObj.MaxBettingNum {
		sendLotteryResponse(w, "0204", strCmd, "非法的抽奖注数", "")
		return
	}

	//扣除积分
	myMsgHandler := &msghelper.MsgHandler{}
	myMsgHandler.SetToken(stNewRequest.Token)
	err := myMsgHandler.AddAccountScore(stNewRequest, deviceID, -lotteryMultiple*myJackpotObj.SingleBettingScore)
	if err != nil {
		//扣除积分失败
		sendLotteryResponse(w, "0205", strCmd, "扣除积分失败", "")
		return
	}


	//if(lotteryType == 0 || lotteryType == 1){
	//	//未蓄力成功，直接返回
	//	strInBody := myMsgHandler.EncodeLotteryResponse(0)

	//	sendLotteryResponse(w, "0100", strCmd, "", strInBody)
	//	return
	//}

	//抽奖
	rewardScore := myJackpotObj.LotteryDraw(lotteryMultiple, lotteryType)

	//增加奖励	
	rnd,_ = rand.Int(rand.Reader,big.NewInt(1000000))
	stNewRequest.OrderId =  strconv.FormatInt(time.Now().Unix(),10) + fmt.Sprintf("%d",rnd)
	err = myMsgHandler.AddAccountScore(stNewRequest, deviceID, rewardScore)
	if err != nil {
		//增加奖励失败
		sendLotteryResponse(w, "0205", strCmd, "增加积分失败", "")
		return
	}

	//处理成功，发送返回
	strRespBody := myMsgHandler.EncodeLotteryResponse(rewardScore)

	sendLotteryResponse(w, "0100", strCmd, "", strRespBody)

	//记录数据库
	strOpenSql := fmt.Sprintf("%s:%s@tcp(%s:3306)/lottery_log?charset=utf8", conf.ServerConfig.DBUser, conf.ServerConfig.DBPasswd, conf.ServerConfig.DBHost)
	db, err := sql.Open("mysql", strOpenSql)
	if err != nil {
		fmt.Println(err)
		return
	}
  	defer db.Close()
  	
	//插入数据
 	stmt, err := db.Prepare("insert lottery_log_tb set cardID=?,userID=?,costScore=?,rewardScore=?,rewardRate=?,opera_time=?")
  	if err != nil {
		fmt.Println(err)
		return
	}
	defer stmt.Close()
  	
	stmt.Exec(stNewRequest.ChipId, stNewRequest.NickId,lotteryMultiple*myJackpotObj.SingleBettingScore,rewardScore,0, time.Now().Format("2006-01-02 15:04:05"))

	return
}

func AdminLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		//拉取登录页面信息
		t, _ := template.ParseFiles("tmpl/login.tmpl")

		//执行解析模版
		t.Execute(w, conf.ServerConfig)
	} else if r.Method == "POST" {
		userName := r.FormValue("username")
		passwd := r.FormValue("password")

		fmt.Printf("username = %s, passwd = %s\n", userName, passwd)		

		//校验参数
		if userName != conf.ServerConfig.UserName || passwd != conf.ServerConfig.Password {
			//参数校验错误
			io.WriteString(w, "Failed to login, invalid user or password!\n")
			return
		}

		//获取session,使用唯一session id
		seesion, _ := cookieStore.Get(r, "session-id")
		//todo jasonxiong 暂时设置为10s
		//seesion.Options.MaxAge = 10
		//seesion.Values["session-id"] = "bar"

		seesion.Save(r, w)

		fmt.Printf("user %s, password %s\n", userName, passwd)

		//登录成功，跳转到配置页面
		http.Redirect(w, r, "/admin/config", http.StatusFound)

		return
	}
}

func AdminConfigHandler(w http.ResponseWriter, r *http.Request) {
	//获取session,使用唯一session id
	seesion, _ := cookieStore.Get(r, "session-id")
	//fmt.Println("sesssion is new:", seesion.IsNew)
	if seesion.IsNew {
		//新建的session，未登录过
		seesion.Options.MaxAge = -10
		http.Redirect(w, r, "/admin/login", http.StatusFound)
		return
	}

	if r.Method == "GET" {
		//拉取配置页面信息
		t, _ := template.ParseFiles("tmpl/config.tmpl")

		//执行解析模版
		t.Execute(w, conf.ServerConfig)
	} else if r.Method == "POST" {
		if len(r.FormValue("reset_param")) != 0{
			
			fmt.Println("Do reset config param")

			//拉取配置页面信息
            t, _ := template.ParseFiles("tmpl/config.tmpl")

            //执行解析模版
            t.Execute(w, conf.ServerConfig)

			return
		} else if len(r.FormValue("gen_solution")) != 0{

			fmt.Println("Do generate solution")

			//生成解决方案
			//获取参数
			myJackpotObj := jackpot.JackpotInstance()
			myJackpotObj.ReturnRate, _ = strconv.Atoi(r.FormValue("returnRate"))
			myJackpotObj.ExpectedRate, _ = strconv.Atoi(r.FormValue("expectedRate"))
			myJackpotObj.SingleBettingScore, _ = strconv.Atoi(r.FormValue("SingleBettingScore"))
			myJackpotObj.MaxBettingNum, _ = strconv.Atoi(r.FormValue("MaxBettingNum"))
			myJackpotObj.RewardDataInfo[1].RewardScore, _ = strconv.Atoi(r.FormValue("RewardScore1"))
			myJackpotObj.RewardDataInfo[2].RewardNum, _ = strconv.Atoi(r.FormValue("RewardNum2"))
			myJackpotObj.RewardDataInfo[2].RewardScore, _ = strconv.Atoi(r.FormValue("RewardScore2"))
			myJackpotObj.RewardDataInfo[3].RewardNum, _ = strconv.Atoi(r.FormValue("RewardNum3"))
			myJackpotObj.RewardDataInfo[3].RewardScore, _ = strconv.Atoi(r.FormValue("RewardScore3"))
			myJackpotObj.RewardDataInfo[4].RewardNum, _ = strconv.Atoi(r.FormValue("RewardNum4"))
			myJackpotObj.RewardDataInfo[4].RewardScore, _ = strconv.Atoi(r.FormValue("RewardScore4"))
			myJackpotObj.RewardDataInfo[5].RewardNum, _ = strconv.Atoi(r.FormValue("RewardNum5"))
			myJackpotObj.RewardDataInfo[5].RewardScore, _ = strconv.Atoi(r.FormValue("RewardScore5"))

			//重新生成奖池
			myJackpotObj.GenerateJackpot()
			conf.ServerConfig.ReturnRate = myJackpotObj.ReturnRate
			conf.ServerConfig.ExpectedRate = myJackpotObj.ExpectedRate
			conf.ServerConfig.SingleBettingScore = myJackpotObj.SingleBettingScore
			conf.ServerConfig.MaxBettingNum = myJackpotObj.MaxBettingNum
			conf.ServerConfig.CurrentWinRate = myJackpotObj.CurrentWinRate
			conf.ServerConfig.RewardNum0 = myJackpotObj.RewardDataInfo[0].RewardNum
			conf.ServerConfig.RewardScore0 = myJackpotObj.RewardDataInfo[0].RewardScore
			conf.ServerConfig.RewardNum1 = myJackpotObj.RewardDataInfo[1].RewardNum
			conf.ServerConfig.RewardScore1 = myJackpotObj.RewardDataInfo[1].RewardScore
			conf.ServerConfig.RewardNum2 = myJackpotObj.RewardDataInfo[2].RewardNum
			conf.ServerConfig.RewardScore2 = myJackpotObj.RewardDataInfo[2].RewardScore
			conf.ServerConfig.RewardNum3 = myJackpotObj.RewardDataInfo[3].RewardNum
			conf.ServerConfig.RewardScore3 = myJackpotObj.RewardDataInfo[3].RewardScore
			conf.ServerConfig.RewardNum4 = myJackpotObj.RewardDataInfo[4].RewardNum
			conf.ServerConfig.RewardScore4 = myJackpotObj.RewardDataInfo[4].RewardScore
			conf.ServerConfig.RewardNum5 = myJackpotObj.RewardDataInfo[5].RewardNum
			conf.ServerConfig.RewardScore5 = myJackpotObj.RewardDataInfo[5].RewardScore
			conf.ServerConfig.RewardNum6 = myJackpotObj.RewardDataInfo[6].RewardNum
			conf.ServerConfig.RewardScore6 = myJackpotObj.RewardDataInfo[6].RewardScore


			//拉取配置页面信息
			t, _ := template.ParseFiles("tmpl/config.tmpl")

			//执行解析模版
			t.Execute(w, conf.ServerConfig)

			return
		} else if len(r.FormValue("confirm")) != 0{
			//确认解决方案

			fmt.Println("Update config to file")

			//更新当前设置到文件中
			conf.UpdateConfig()
			
			//拉取配置页面信息
            t, _ := template.ParseFiles("tmpl/config.tmpl")

            //执行解析模版
            t.Execute(w, conf.ServerConfig)

			return

		} else{

			return
		}


		return
	}
}

func AdminQueryHandler(w http.ResponseWriter, r *http.Request){
	if r.Method == "GET"{
		//拉取查询页面信息
		t,_ := template.ParseFiles("tmpl/query.tmpl")

		//执行解析模板
		t.Execute(w, conf.ServerConfig)

		return
	} else if r.Method == "POST"{

		//获取请求消息
		//strRequest, _ := ioutil.ReadAll(r.Body)
		//fmt.Printf("client request: %s\n", strRequest)		

		//r.ParseForm()

		fmt.Println(r.FormValue("type"))	

		//获取参数
		querytype,_ := strconv.Atoi(r.FormValue("type"))
		queryid,_ := strconv.Atoi(r.FormValue("queryid"))
		begindate := r.FormValue("begindate")
		enddate := r.FormValue("enddate")
	
		fmt.Printf("type: %s, id: %s, begindate:%s, enddate:%s\n", r.FormValue("type"), r.FormValue("queryid"), begindate, enddate)
			
		//查询条件
		strCondition := ""
		switch querytype{
		case 1:
			strCondition = " "	
		case 2:
			strCondition = fmt.Sprintf(" cardID='%s' and ", queryid)
		case 3:
			strCondition = fmt.Sprintf(" userID='%s' and ", queryid)
		default:
			return
		}	

		//生成查询语句
		strQuery := fmt.Sprintf("select * from lottery_log_tb where %s opera_time>='%s' and opera_time<='%s'", strCondition, begindate, enddate)

		fmt.Println("query string : ", strQuery)		

		//执行查询
		strOpenSql := fmt.Sprintf("%s:%s@tcp(%s:3306)/lottery_log?charset=utf8", conf.ServerConfig.DBUser, conf.ServerConfig.DBPasswd, conf.ServerConfig.DBHost)
		db, err := sql.Open("mysql", strOpenSql)
		if err != nil {
			fmt.Println(err)
			return
		}
  		defer db.Close()

		rows, err := db.Query(strQuery)
		if err != nil {
			fmt.Println(err)
			return
		}
 		defer rows.Close()
		
		var respdata []map[string]interface{}

		var cardID string
		var userID string
		var costScore int
		var rewardScore int
		var rewardRate int
		var opera_time string
		for rows.Next() {
			err := rows.Scan(&cardID, &userID, &costScore, &rewardScore, &rewardRate, &opera_time)
			if err != nil {
				fmt.Println(err)
				continue
			}
			
			arrData := make(map[string]interface{})
			arrData["cardID"] = cardID
			arrData["userID"] = userID
			arrData["costScore"] = costScore
			arrData["rewardScore"] = rewardScore
			arrData["rewardRate"] = rewardRate
			arrData["opera_time"] = opera_time
			respdata = append(respdata, arrData)			
			
		}
 
		err = rows.Err()
		if err != nil {
			fmt.Println(err)
			return
		}	

		//封装返回
		strResp := make(map[string]interface{})
		strResp["data"] = respdata

		binData, _ := json.Marshal(strResp)

		io.WriteString(w, string(binData))

		return
	}
}

func getSessionID() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}

	return base64.URLEncoding.EncodeToString(b)
}

func sendLotteryResponse(w http.ResponseWriter, result, cmd, msg, body string) {
	respjs, _ := simplejson.NewJson([]byte(`{}`))

	//先封装消息头
	respjs.SetPath([]string{"header", "resultStatus"}, result)
	respjs.SetPath([]string{"header", "cmd"}, cmd)
	respjs.SetPath([]string{"header", "msg"}, msg)

	if result == "0100" {
		//成功的消息，封装消息体
		respjs.Set("body", body)
	}

	//json编码
	strresp, _ := respjs.Encode()

	io.WriteString(w, string(strresp))

	return
}
