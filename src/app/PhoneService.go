package app

import (
	"SmsBpi/utils"
	"SmsBpi/config"
	"time"
	"encoding/hex"
	"strings"
	"sync"
)

var wLock sync.Mutex

var taskBus chan utils.PhoneCmd = make(chan utils.PhoneCmd, 100)
var resultBus chan utils.PhoneCmd = make(chan utils.PhoneCmd, 100)

func listenSms() {
	// 每间隔5s，向总线发送一条检查未读短信的指令
	for {
		wLock.Lock()
		utils.ListenSms(taskBus)
		wLock.Unlock()
		time.Sleep(time.Duration(5) * time.Second)
	}
}

func clearSms() {
	// 每天清理一次已读短信
	for {
		wLock.Lock()
		utils.CleanSms(taskBus)
		wLock.Unlock()
		time.Sleep(time.Duration(24) * time.Hour)
	}
}

func heartBeat() {
	// 每10s进行一次信号强度上报
}

// 处理AT命令执行结果
func ProcessATcmdResult(result chan utils.PhoneCmd, config *config.Config, token *string) {
	for {
		phoneMsg := <-result
		// 可以在这里对不同指令的处理结果
		if strings.HasPrefix(string(phoneMsg.ATCmd), utils.CMD_CMGL_ALL) && strings.Contains(phoneMsg.Result, "OK") {
			msgs := strings.Split(phoneMsg.Result, utils.CMD_LF_CR)
			for i, m := range msgs {
				if strings.HasPrefix(m, "+CMGL:") && strings.Contains(m, "UNREAD") {
					tmp_info := strings.Split(msgs[i], ",\"")
					src := strings.ReplaceAll(tmp_info[2], "\"", "")
					src = strings.ReplaceAll(src, ",", "")
					tmpsrc, _ := hex.DecodeString(src)
					phonenum, _ := utils.Ucs2ToUtf8(string(tmpsrc))
					t := strings.Replace(tmp_info[3], "\"", "", -1) //SIM900A 格式应改为：tmp_info[4]
					if len(phoneMsg.SendMSG) != 0 {
						phoneMsg.SendMSG += "\n"
					}
					phoneMsg.SendMSG += "来源: " + phonenum + " 时间: " + t + "\n"
					if utils.IsUcs(msgs[i+1]) {
						dat, _ := hex.DecodeString(msgs[i+1])
						tmpmsg, _ := utils.Ucs2ToUtf8(string(dat))
						phoneMsg.SendMSG += tmpmsg
					} else {
						phoneMsg.SendMSG += msgs[i+1]
					}
				}
			}
		}
		if strings.HasPrefix(string(phoneMsg.ATCmd), utils.CMD_ATD) {
			if strings.Contains(phoneMsg.Result, "OK") {
				phoneMsg.SendMSG = "拨打电话成功"
			} else {
				phoneMsg.SendMSG = "拨打电话失败"
			}
		}

		if strings.HasPrefix(string(phoneMsg.ATCmd), utils.CMD_CMGS) {
			if strings.Contains(phoneMsg.Result, "ERROR") {
				phoneMsg.SendMSG = "发送短信失败"
			} else {
				phoneMsg.SendMSG = "发送短信成功"
			}
		}
	}
}

func Run(config config.Config) {
	go listenSms()
	go heartBeat()
	go utils.ExecATCmd(taskBus, resultBus, &config)
	// go processATcmdResult()
}
