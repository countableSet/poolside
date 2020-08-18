package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type Configuration struct {
	Domain string `json:"domain"`
	Proxy  string `json:"proxy"`
}

var ConfigUpdateChan = make(chan []Configuration)

func writeToFile(c *[]Configuration) {
	j, _ := json.Marshal(c)
	_ = ioutil.WriteFile("config.json", j, 0644)
}

func readFromFile() []Configuration {
	var c []Configuration
	f, err := os.Open("config.json")
	if err != nil {
		log.Printf("Error opening configuration file %v", err)
		return []Configuration{}
	}
	defer f.Close()
	b, _ := ioutil.ReadAll(f)
	json.Unmarshal(b, &c)
	log.Printf("Configuration read: %v", c)
	return c
}