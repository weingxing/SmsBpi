package utils

import (
	"strings"
	"time"
	"log"
	"github.com/jacobsa/go-serial/serial"
	"SmsBpi/config"
)

const (
	CMD_CSQ       string = "AT+CSQ"  // 信号强度
	CMD_CPIN      string = "AT+CPIN?" // 检测是否存在SIM卡
	CMD_CREG      string = "AT+CREG?" // 检测是否注册网络
	CMD_CGATT     string = "AT+CGATT?"  // 检测是否附着GPRS
	CMD_CPMS      string = "AT+CPMS?"  // 查询SIM卡短信使用情况
	CMD_COPS      string = "AT+COPS?"
	CMD_GSN       string = "AT+CMD_GSN"  // 查询IMEI
	CMD_CCID      string = "AT+CCID"  // 查询SIM卡CCID
	CMD_CMGF      string = "AT+CMGF=1"  // 设置为文本模式
	CMD_CSCS_UCS2 string = "AT+CSCS=\"UCS2\"" //设置编码（中文短信）
	CMD_CSCS_GSM  string = "AT+CSCS=\"GSM\""  //设置编码（英文短信）
	CMD_CSMP      string = "AT+CSMP=17,71,0,8" //17,167,0,8 17,167,2,25  1.文本模式，2.this， 3.UCS2， 4.手机号Unicode，5.内容，6.0x1A
	CMD_CMGFZ     string = "AT+CMGF=0"
	CMD_CMGL_ALL  string = "AT+CMGL=\"REC UNREAD\"" //获取所有未读短信
	CMD_CMGDA_ALL string = "AT+CMGDA=\"DEL ALL\""            // 删除已读短信
	CMD_CMGS      string = "AT+CMGS=\""             //发送短信指令 后跟手机号码
	CMD_ATD       string = "ATD"                    //呼叫号码
	CMD_ATH       string = "ATH"                    //挂机
	CMD_CTRL_Z    string = "\x1A"   // 发送短信指令
	CMD_LF_CR     string = "\r\n"
	CMD_LF        string = "\r"
	CACHE_SIZE    int    = 1024 * 8
)

type SendMsgResp struct {
	ErrorCode   int    `json:"errorcode"`
	Errmsg      string `json:"errmsg"`
	Invaliduser string `json:"invaliduser"`
}

type PhoneCmd struct {
	ATCmd     []byte
	CmdDirect string
	Result    string
	SendMSG   string
	Delay  uint
}

func CheckErr(err error) {
	if err != nil {
		log.Fatalf(err.Error())
	}
}

// 执行AT命令
func ExecATCmd(input chan PhoneCmd, result chan PhoneCmd, config *config.Config) {
	options := serial.OpenOptions{
		PortName:        config.Device,
		BaudRate:        config.Baudrate,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}

	port, err := serial.Open(options)
	CheckErr(err)
	defer port.Close()
	info_cache := make([]byte, CACHE_SIZE)
	for {
		cmd := <-input
		if strings.HasPrefix(string(cmd.ATCmd[:]), CMD_CMGS) {
			bodys := strings.SplitN(string(cmd.ATCmd[:]), ":::", 2)
			tmp_control := string(bodys[0]) + CMD_LF
			_, err = port.Write([]byte(tmp_control))
			time.Sleep(time.Duration(1) * time.Second)
			CheckErr(err)
			port.Write([]byte(bodys[1]))
			info_cache = make([]byte, CACHE_SIZE)
			port.Read(info_cache)
			cmd.Result = string(info_cache[:])
			result <- cmd
		} else {
			_, err = port.Write(cmd.ATCmd)
			CheckErr(err)
			time.Sleep(time.Duration(1) * time.Second)
			port.Read(info_cache)
			cmd.Result = string(info_cache[:])
			if strings.HasPrefix(cmd.Result, ">") {
				port.Write([]byte(CMD_CTRL_Z))
			}
			result <- cmd
			info_cache = make([]byte, CACHE_SIZE)
		}
	}
}

func SendSms(taskBus chan PhoneCmd, msg string, phone string) {
	phoneCode := GetStrUnicode(phone)
	smsBody := GetStrUnicode(msg)

	// AT+CMGF=1
	setMessageFormatCmd := []byte(CMD_CMGF + CMD_LF_CR)
	cmd := PhoneCmd{Delay: 1}
	cmd.ATCmd = setMessageFormatCmd
	taskBus <- cmd
	
	ucs2cmd := []byte(CMD_CSCS_UCS2 + CMD_LF_CR)
	setucs2 := PhoneCmd{Delay: 1}
	setucs2.ATCmd = ucs2cmd
	taskBus <- setucs2

	csmpcmd := []byte(CMD_CSMP + CMD_LF_CR)
	setcsmp := PhoneCmd{Delay: 1}
	setcsmp.ATCmd = csmpcmd
	taskBus <- setcsmp

	phoneMsg := PhoneCmd{Delay: 2}
	phoneMsg.ATCmd = []byte(CMD_CMGS + phoneCode + "\":::" + smsBody + CMD_CTRL_Z)
	taskBus <- phoneMsg
}

func ListenSms(taskBus chan PhoneCmd) {
	readAllMessagesCmd := []byte(CMD_CMGL_ALL + CMD_LF_CR)
	cmd := PhoneCmd{Delay: 1}
	cmd.ATCmd = readAllMessagesCmd
	taskBus <- cmd
}

func CleanSms(taskBus chan PhoneCmd) {
	cleanAllReadMessagesCmd := []byte(CMD_CMGDA_ALL + CMD_LF_CR)
	cmd := PhoneCmd{Delay: 1}
	cmd.ATCmd = cleanAllReadMessagesCmd
	taskBus <- cmd
}

func SignalStrength(taskBus chan PhoneCmd) {
	signalStrengthCmd := []byte(CMD_CSQ + CMD_LF_CR)
	cmd := PhoneCmd{Delay: 1}
	cmd.ATCmd = signalStrengthCmd
	taskBus <- cmd
}

func Operator(taskBus chan PhoneCmd) {
	operatorCmd := []byte(CMD_COPS + CMD_LF_CR)
	cmd := PhoneCmd{Delay: 1}
	cmd.ATCmd = operatorCmd
	taskBus <- cmd
}
