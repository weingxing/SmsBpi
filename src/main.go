package main

import (
	"SmsBpi/config"
	"SmsBpi/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"github.com/jacobsa/go-serial/serial"
	"strings"
)

func loadConfig(file string, cfg *config.Config) {
	file_body, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(file_body, cfg)
}

func main() {
	var cfg config.Config
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s config.json\n", os.Args[0])
		return
	}
	loadConfig(os.Args[1], &cfg)
	// utils.Bark("HelloWorld", cfg)
	 
// AT+CMGF=1
// AT+CSCS="UCS2"
// AT+CSMP=17,167,0,8          //表示普通文本模式
// AT+CMGS="00310030003000380036"  
// > 0031     //短信内容
// 1A              //表示发送
	// fmt.Println(utils.GetStrUnicode("10086"))
	// fmt.Println(utils.GetStrUnicode("1"))
	// fmt.Println(utils.Utf8ToUcs2("你好"))
	// fmt.Println(utils.Ucs2ToUtf8("4F60597D"))
	
	options := serial.OpenOptions{
		PortName:        "COM3",
		BaudRate:        115200,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}

	port, err := serial.Open(options)
	if err != nil {
		fmt.Println(err)
	}
	port.Write([]byte("AT+CMGR=5\r\n"))
	info_cache := make([]byte, 1024 * 8)
	port.Read(info_cache)
	msg := string(info_cache[:])
	port.Close()
	a := strings.Split(msg, ",")
	// b := "00310030003600380035003700370037003600300035003000370034"
	fmt.Println(utils.DecodeUcs2(strings.ReplaceAll(a[1], "\"", "")))
	fmt.Println(strings.ReplaceAll("20" + a[3] + " " + strings.Split(a[4], "\n")[0], "\"", ""))
	fmt.Println(utils.DecodeUcs2(strings.Split(a[4], "\n")[1]))
}

