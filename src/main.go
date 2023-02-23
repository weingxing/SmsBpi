package main

import (
	"SmsBpi/config"
	"SmsBpi/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
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
	utils.Bark("HelloWorld", cfg)
}
