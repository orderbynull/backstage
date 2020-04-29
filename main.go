package main

import (
	"backstage/pgsql"
	"encoding/json"
	"fmt"
	protocol "github.com/orderbynull/protocol/pgsql"
	"os"
	"regexp"
)

func main() {
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "init" {
		cfg := Config{
			[]Proxy{
				{
					Type:       "pgsql",
					SourceName: "PHP_APP",
					TargetName: "PGSQL",
					ListenAddr: "127.0.0.1:4046",
					TargetAddr: "127.0.0.1:5432",
				},
			},
		}

		jsonData, _ := json.MarshalIndent(cfg, "", "  ")

		config, _ := os.Create("./config.json")
		config.Write(jsonData)

		os.Exit(0)
	}

	config := readConfig("./config.json")

	space := regexp.MustCompile(`\s+`)

	for _, proxy := range config.Proxies {
		switch proxy.Type {
		case "pgsql":
			messages := pgsql.NewProxy(proxy.ListenAddr, proxy.TargetAddr).Run()
			go func(messages chan interface{}, proxy Proxy) {
				for m := range messages {
					switch msg := m.(type) {
					case protocol.ParseMessage:
						if len(msg.String()) > 0 {
							fmt.Printf("[%s@%s]> %s\n\n", proxy.SourceName, proxy.TargetName, space.ReplaceAllString(msg.String(), " "))
						}
					}
				}
			}(messages, proxy)
		}
	}

	<-make(chan interface{})
}
