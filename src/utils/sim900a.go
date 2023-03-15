package utils

import (
	// "SmsBpi/config"
	"log"
	// "strings"
	// "time"
)

const (
	CMD_ATE0      string = "ATE0"                   // 关闭命令回显
	CMD_CSQ       string = "AT+CSQ"                 // 信号强度
	CMD_CPIN      string = "AT+CPIN?"               // 检测是否存在SIM卡
	CMD_CREG      string = "AT+CREG?"               // 检测是否注册网络
	CMD_CGATT     string = "AT+CGATT?"              // 检测是否附着GPRS
	CMD_CGATT_OFF string = "AT+CGATT=0"             // 分离GPRS网络
	CMS_CGAT_ON   string = "AT+CGATT=1"             // 附着GPRS网络
	CMD_CPMS      string = "AT+CPMS?"               // 查询SIM卡短信使用情况
	CMD_CNMI      string = "AT+CNMI=2,2,0,0,0"      // 短信直接显示，不存储到SIM卡
	CMD_COPS      string = "AT+COPS?"               // 查询运营商
	CMD_GSN       string = "AT+CMD_GSN"             // 查询IMEI
	CMD_CCID      string = "AT+CCID"                // 查询SIM卡CCID
	CMD_CMGF      string = "AT+CMGF=1"              // 设置短信为文本模式
	CMD_CMGF0     string = "AT+CMGF=0"              // 设置短信为PDU模式
	CMD_CSCS_UCS2 string = "AT+CSCS=\"UCS2\""       // 设置编码（中文短信）
	CMD_CSCS_GSM  string = "AT+CSCS=\"GSM\""        // 设置编码（英文短信）
	CMD_CSMP      string = "AT+CSMP=17,167,0,8"     // 文本模式
	CMD_CMGFZ     string = "AT+CMGF=0"              // 设置短信为PDU模式
	CMD_CMGL_ALL  string = "AT+CMGL=\"REC UNREAD\"" // 获取所有未读短信
	CMD_CMGDA_ALL string = "AT+CMGDA=\"DEL ALL\""   // 删除已读短信
	CMD_CMGS      string = "AT+CMGS=\""             // 发送短信指令 后跟手机号码
	CMD_ATD       string = "ATD"                    // 呼叫号码
	CMD_ATH       string = "ATH"                    // 挂机
	CMD_ATA       string = "ATA"                    // 接听
	CMD_CTRL_Z    string = "\x1A"                   // 发送短信指令（十六进制）
	CMD_LF_CR     string = "\r\n"                   // 回车换行
	CMD_LF        string = "\r"                     // 回车
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
	Delay     uint
}

func CheckErr(err error) {
	if err != nil {
		log.Fatalf(err.Error())
	}
}

func SendSms(taskBus chan PhoneCmd, phone string, msg string) {
	phoneCode := EncodeUcs2(phone)
	smsBody := EncodeUcs2(msg)

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

	phoneMsg := PhoneCmd{Delay: 5}
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

func Initlize(taskBus chan PhoneCmd) {
	ate0 := []byte(CMD_ATE0 + CMD_LF_CR)
	ateCmd := PhoneCmd{Delay: 1}
	ateCmd.ATCmd = ate0
	taskBus <- ateCmd // 关闭回显
	gprsOff := []byte(CMD_CGATT_OFF + CMD_LF_CR)
	gprsOffCmd := PhoneCmd{Delay: 1}
	gprsOffCmd.ATCmd = gprsOff
	taskBus <- gprsOffCmd // 脱离GPRS，防止没有流量导致额外费用
	sms := []byte(CMD_CNMI + CMD_LF_CR)
	smsCmd := PhoneCmd{Delay: 1}
	smsCmd.ATCmd = sms // 短信直接串口传输，不存储到SIM卡
	taskBus <- smsCmd
}
