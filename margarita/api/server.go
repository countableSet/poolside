package api

import (
	"encoding/json"
	"log"
	"net/http"
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

	log.Printf("api server listening %d", 3000)
	http.ListenAndServe(":3000", nil)
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
