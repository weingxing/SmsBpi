package app

import (
	"SmsBpi/config"
	"SmsBpi/utils"
	"fmt"
	"github.com/jacobsa/go-serial/serial"
	"io"
	"log"
	"strings"
	"sync"
	"time"
)

var wLock sync.Mutex

var taskBus chan utils.PhoneCmd = make(chan utils.PhoneCmd, 100)
var resultBus chan utils.PhoneCmd = make(chan utils.PhoneCmd, 100)
var msgBus chan []byte = make(chan []byte, 100)

var port io.ReadWriteCloser = nil

const BUFFER_SIZE = 1024 * 8

func connDevice(config config.Config) {
	options := serial.OpenOptions{
		PortName:        config.Device,
		BaudRate:        config.BaudRate,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}
	p, err := serial.Open(options)
	if err != nil {
		log.Println(err)
	}
	port = p
}

func close() {
	if port != nil {
		port.Close()
	}
}

func initlizing() {
	// 检查SIM卡状态，正常后进行初始化
	log.Println("Init")
	wLock.Lock()
	utils.Initlize(taskBus)
	wLock.Unlock()
}

func listenSms(config config.Config) {
	buffer := make([]byte, BUFFER_SIZE)
	for {
		port.Read(buffer)
		// todo handle buffer
		msg := string(buffer[:])
		if strings.Contains(msg, "+CMT:") {
			// 存储短信，Bark
			// 发送成功 +CMGS: 157
			log.Println("来短信了")
			msgBodys := strings.Split(msg, ",")
			fmt.Println(msgBodys)
			// log.Println(utils.DecodeUcs2(strings.ReplaceAll(msgBodys[1], "\"", "")))
			// log.Println()
			// receiveTime := strings.ReplaceAll("20" + msgBodys[3] + " " + strings.Split(msgBodys[4], "\n")[0], "\"", "")
			// body := utils.DecodeUcs2(strings.Split(msgBodys[4], "\n")[1])
			// utils.Bark(receiveTime + "\n" + body, config)
			// log.Println(utils.DecodeUcs2(b))
			// log.Println(utils.EncodeUcs2("10685777605074"))
			// log.Println(b)
		} else {
			// 放回结果队列
			// log.Println(msg)
			msgBus <- buffer[:]
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
			_, err := port.Write([]byte(tmpControl))
			if err != nil {
				log.Fatal(err)
			}
			// 发送消息内容和确认发送标志
			time.Sleep(time.Duration(1) * time.Second)
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
func processATcmdResult(config config.Config) {
	for {
		msg := <-msgBus
		cmd := <-resultBus
		body := string(msg)
		if strings.Contains(body, "ERROR") {
			utils.Bark(string(cmd.ATCmd)+"执行失败", config)
		} else if strings.Contains(body, "+CSQ") {
			strength := strings.Split(body, " ")[1]
			s := strings.Split(strength, "\n")[0]
			signal := strings.Split(s, ",")[0]
			log.Println("信号强度: -" + signal + "dBM")
		} else {
			log.Println(body)
		}
		// todo 判定短信是否发送成功
		// todo 处理信号强度
		// todo 处理运营商
		// todo 处理IMEI
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
	go processATcmdResult(config)
	go execATCmd()
	go listenSms(config)
	go heartBeat()
	SendSmS("10086", "1")
	wg.Wait()
}
