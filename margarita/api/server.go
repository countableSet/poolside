package api

import (
	"encoding/json"
	"log"
	"net/http"
)

func RunApiServer() {
	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/api/configurations", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGet(w, r)
		case http.MethodPost:
			handlePost(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	log.Printf("api server listening %d", 3000)
	http.ListenAndServe(":3000", nil)
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	// sample data for now
	w.Header().Set("Content-Type", "application/json")
	payload := []Configuration{
		{Domain: "test.local.bimmer-tech.com", Proxy: "localhost:8000"},
	}
	_ = json.NewEncoder(w).Encode(payload)
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
	log.Printf("%v", c)
	w.WriteHeader(http.StatusCreated)
}