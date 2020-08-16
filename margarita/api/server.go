package api

import (
	"log"
	"net/http"
)

func RunApiServer() {
	http.Handle("/", http.FileServer(http.Dir("./public")))

	log.Printf("api server listening %d", 3000)
	http.ListenAndServe(":3000", nil)
}
