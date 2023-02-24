package main

import (
	"SmsBpi/app"
	"SmsBpi/config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	// "github.com/jacobsa/go-serial/serial"
	// "strings"
)

func loadConfig(file string, cfg *config.Config) {
	file_body, err := ioutil.ReadFile(file)
	if err != nil {
		os.Exit(-1)
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
	app.Run(cfg)
}
