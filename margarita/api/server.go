package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/countableset/poolside/margarita/config"
)

var configs = readFromFile()

func RunApiServer() {
	ConfigUpdateChan <- configs
	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/api/configurations", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGet(w)
		case http.MethodPost:
			handlePost(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	host := config.GetMargaritaHost()
	port := config.GetMargaritaPort()
	log.Printf("api server listening %s:%d", host, port)
	http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil)
}

func handleGet(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(configs)
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	var c []Configuration
	if err := decoder.Decode(&c); err != nil {
		log.Printf("error decoding request body %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("Decoded and write back to file: %v", c)
	configs = c
	writeToFile(&c)
	ConfigUpdateChan <- configs
	w.WriteHeader(http.StatusCreated)
}
