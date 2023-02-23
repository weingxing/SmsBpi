package utils

import (
	"SmsBpi/config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
)

func Bark(msg string, cfg config.Config) bool {
	url := fmt.Sprintf("%s/%s/%s", cfg.BarkServer, cfg.BarkSecret, msg)
	fmt.Println(url)
	resp, err := http.Get(url)
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
