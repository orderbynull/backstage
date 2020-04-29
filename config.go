package main

import (
	"encoding/json"
	"log"
	"os"
)

type Proxy struct {
	Type       string `json:"type"`
	SourceName string `json:"sourceName"`
	TargetName string `json:"targetName"`
	ListenAddr string `json:"listenAddr"`
	TargetAddr string `json:"targetAddr"`
}

type Config struct {
	Proxies []Proxy `json:"proxies"`
}

func readConfig(filename string) Config {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := Config{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalln(err)
	}

	return config
}
