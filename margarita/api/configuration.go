package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

// Configuration structure for storing proxy information
type Configuration struct {
	Domain string `json:"domain"`
	Proxy  string `json:"proxy"`
}

var (
	// ConfigUpdateChan channel to listener for configuratin updates
	ConfigUpdateChan = make(chan []Configuration)
	configPath       = os.Getenv("CONFIG_PATH") + "/config.json"
)

func writeToFile(c *[]Configuration) {
	j, _ := json.Marshal(c)
	_ = ioutil.WriteFile(configPath, j, 0644)
}

func readFromFile() []Configuration {
	var c []Configuration
	f, err := os.Open(configPath)
	if err != nil {
		log.Printf("Using empty configuration, error with file: %v", err)
		return []Configuration{}
	}
	defer f.Close()
	b, _ := ioutil.ReadAll(f)
	json.Unmarshal(b, &c)
	log.Printf("Configuration read: %v", c)
	return c
}
