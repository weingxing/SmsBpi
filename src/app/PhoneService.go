package app

import (
	"SmsBpi/config"
	"SmsBpi/utils"
	"go.bug.st/serial"
	"log"
	"strings"
	"sync"
	"time"
)

var wLock sync.Mutex

var taskBus chan utils.PhoneCmd = make(chan utils.PhoneCmd, 100)
var resultBus chan utils.PhoneCmd = make(chan utils.PhoneCmd, 100)
var msgBus chan []byte = make(chan []byte, 100)
var to string // 短信收信人

var port serial.Port = nil

const BUFFER_SIZE = 1024 * 8

func connDevice(config config.Config) {
	options := &serial.Mode{
		BaudRate: 115200,
	}
	p, err := serial.Open(config.Device, options)
	if err != nil {
		log.Fatal(err)
	}
	port = p
}

func close() {
	if port != nil {
		port.Close()
	}
}

func initlizing() {
	buf := make([]byte, BUFFER_SIZE)
	// 检查SIM卡状态，正常后进行初始化
	ready := false
	for !ready {
		port.Write([]byte("AT\r\n"))
		time.Sleep(time.Duration(1) * time.Second)
		port.Read(buf)
		if strings.Contains(string(buf), "OK") {
			ready = true
			log.Println("串口通信正常")
		}
	}
	ready = false
	for !ready {
		port.Write([]byte("AT+CPIN?\r\n"))
		time.Sleep(time.Duration(1) * time.Second)
		port.Read(buf)
		if strings.Contains(string(buf), "READY") {
			ready = true
			log.Println("SIM卡状态正常")
		}
	}
	wLock.Lock()
	utils.Initlize(taskBus)
	wLock.Unlock()
}

func listenSms(config config.Config) {
	buffer := make([]byte, BUFFER_SIZE)
	for {
		n, _ := port.Read(buffer)
		sms := string(buffer[:])
		if strings.Contains(sms, "+CMT:") {
			// 存储短信&Bark
			log.Println("来短信了")
			// 等待数据传输完成
			time.Sleep(time.Duration(1) * time.Second)
			// 再次读取，取得全部数据
			buf := make([]byte, BUFFER_SIZE)
			size, _ := port.Read(buf)
			// 拼接数据
			buffer = append(buffer[:n], buf[:size]...)
			bodys := strings.Split(string(buffer[:]), ",")
			sender := utils.DecodeUcs2(strings.Split(bodys[0], "\"")[1])
			receiveTime := "20" + strings.ReplaceAll(bodys[2], "\"", "")
			t := strings.Split(bodys[3], "\r\n")
			receiveTime += strings.ReplaceAll(" "+t[0], "\"", "")
			receiveTime = strings.SplitN(receiveTime, "+", 2)[0]
			body := utils.DecodeUcs2(t[1])
			utils.Bark(sender, receiveTime+"\n"+body, config)
			buffer = make([]byte, BUFFER_SIZE)
		} else if len(strings.ReplaceAll(string(buffer), string(0), "")) > 0 {
			//	// 放到队列
			msgBus <- buffer[:]
			buffer = make([]byte, BUFFER_SIZE)
		}
	}
}

func heartBeat() {
	// 每10s进行一次信号强度上报
	for {
		wLock.Lock()
		utils.SignalStrength(taskBus)
		wLock.Unlock()
		time.Sleep(time.Duration(10) * time.Second)
	}
}

// 执行AT命令
func execATCmd() {
	for {
		cmd := <-taskBus
		// 发送短信，由于是多个指令，需要特殊处理
		if strings.HasPrefix(string(cmd.ATCmd[:]), utils.CMD_CMGS) {
			bodys := strings.SplitN(string(cmd.ATCmd[:]), ":::", 2)
			// 正式发送前的最后一条指令
			tmpControl := string(bodys[0]) + utils.CMD_LF
			resultBus <- utils.PhoneCmd{ATCmd: []byte(tmpControl)}
			_, err := port.Write([]byte(tmpControl))
			if err != nil {
				log.Fatal(err)
			}
			// 发送消息内容和确认发送标志
			time.Sleep(time.Duration(1) * time.Second)
			resultBus <- utils.PhoneCmd{ATCmd: []byte(bodys[1])}
			port.Write([]byte(bodys[1]))
		} else {
			_, err := port.Write(cmd.ATCmd)
			if err != nil {
				log.Fatal(err)
			}
		}
		resultBus <- cmd
		// 每次执行一条指令后，等待给定时间
		time.Sleep(time.Duration(cmd.Delay) * time.Second)
	}
}

// 处理AT命令执行结果，指令执行失败时发出通知
func processATCmdResult(config config.Config) {
	for {
		message := <-msgBus
		cmd := <-resultBus
		body := string(message)
		if strings.Contains(body, "ERROR") {
			utils.Bark("命令执行失败", string(cmd.ATCmd)+"执行失败", config)
		} else if strings.Contains(body, "+CSQ") {
			strength := strings.Split(body, " ")[1]
			s := strings.Split(strength, "\n")[0]
			signal := strings.Split(s, ",")[0]
			log.Println("信号强度: -" + signal + "dBM")
		} else if strings.Contains(body, ">") {
			// 短信接收人
			to = utils.DecodeUcs2(strings.Split(string(cmd.ATCmd), "\"")[1])
		} else if strings.Contains(body, "+CMGS") {
			if !strings.Contains(body, "ERROR") {
				log.Println("发给"+to+"的短信：",
					utils.DecodeUcs2(strings.ReplaceAll(string(cmd.ATCmd), "\x1A", "")))
				log.Println("发送成功")
				utils.Bark("发给："+to, "短信发送成功", config)
			} else {
				utils.Bark("发给："+to, "短信发送失败", config)
			}
		} else if strings.Contains(body, "OK") {
			// ignore
		} else {
			log.Println(cmd, body)
		}
	}
}

func SendSmS(phone string, msg string) {
	wLock.Lock()
	utils.SendSms(taskBus, phone, msg)
	wLock.Unlock()
}

func Run(config config.Config) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	connDevice(config)
	defer close()
	initlizing()
	go processATCmdResult(config)
	go execATCmd()
	go listenSms(config)
	go ListenSend(config)
	//go heartBeat()
	wg.Wait()
}
