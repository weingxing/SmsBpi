package app

import (
	"SmsBpi/config"
	"encoding/json"
	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
	"log"
	"os"
	"os/signal"
)

type MsgBody struct {
	Phone string `json:"phone"`
	Msg   string `json:"msg"`
}

func listenSend(config config.Config) {
	// Set up channel on which to send signal notifications.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill)

	// Create an MQTT Client.
	cli := client.New(&client.Options{
		// Define the processing of the error handler.
		ErrorHandler: func(err error) {
			log.Fatal("create client err: ", err)
		},
	})

	// Terminate the Client.
	defer cli.Terminate()
	// Connect to the MQTT Server.
	err := cli.Connect(&client.ConnectOptions{
		Network:  "tcp",
		Address:  config.MqAddress,
		ClientID: []byte("SmsBpi"),
	})
	if err != nil {
		log.Fatal("connection err: ", err)
	}

	// Subscribe to topics.
	err = cli.Subscribe(&client.SubscribeOptions{
		SubReqs: []*client.SubReq{
			&client.SubReq{
				TopicFilter: []byte(config.Topic),
				QoS:         mqtt.QoS0,
				Handler: func(topicName, message []byte) {
					// 发送短信
					var data MsgBody
					err := json.Unmarshal(message, &data)
					if err != nil {
						log.Fatal("wrong data: ", err)
					}
					log.Println("发送短信: ", string(message))
					SendSmS(data.Phone, data.Msg)
				},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	<-sigc
	// Disconnect the Network Connection.
	if err := cli.Disconnect(); err != nil {
		log.Fatal(err)
	}
}
