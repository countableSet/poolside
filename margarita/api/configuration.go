package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type configuration struct {
	Domain string `json:"domain"`
	Proxy  string `json:"proxy"`
}

func writeToFile(c *[]configuration) {
	j, _ := json.Marshal(c)
	_ = ioutil.WriteFile("config.json", j, 0644)
}

func readFromFile() []configuration {
	var c []configuration
	f, err := os.Open("config.json")
	if err != nil {
		log.Printf("Error opening configuration file %v", err)
		return []configuration{}
	}
	defer f.Close()
	b, _ := ioutil.ReadAll(f)
	json.Unmarshal(b, &c)
	log.Printf("Configuration read: %v", c)
	return c
}