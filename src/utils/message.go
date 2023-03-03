package utils

import (
	"SmsBpi/config"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
)

func Bark(title string, msg string, cfg config.Config) bool {
	url := fmt.Sprintf("%s/%s", cfg.BarkServer, cfg.BarkSecret)
	paramMap := map[string]interface{}{}
	paramMap["title"] = title
	paramMap["body"] = msg
	paramJson, _ := json.Marshal(paramMap)
	param := bytes.NewBuffer([]byte(paramJson))
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, param)
	if err != nil {
		fmt.Println(err)
		return false
	}
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var res map[string]interface{}
	json.Unmarshal(body, &res)
	return res["message"] == "success"
}

func SendEmail(subject string, body string, cfg config.Config) bool {
	from := cfg.SmtpUser
	passwd := cfg.SmtpPassword
	to := cfg.Email

	msg := "From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n\r\n" +
		body

	addr := fmt.Sprintf("%s:%d", cfg.SmtpServer, cfg.SmtpPort)
	err := smtp.SendMail(addr, smtp.PlainAuth("", from, passwd, cfg.SmtpServer), from, []string{to}, []byte(msg))
	if err != nil {
		log.Fatal(err)
	}
	return true
}
