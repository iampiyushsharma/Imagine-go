package main

import (
	"fmt"
	"log"
	"net/http"
	"server/router"
)

func enableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	fmt.Println("MongoDB API")
	r := router.Router()
	corsRouter := enableCors(r)
 godotenv.Load()
 port := os.Getenv('PORT') || ":4000"
	fmt.Println("Server is getting started...")
	log.Fatal(http.ListenAndServe(port, corsRouter))
	fmt.Println("Listening at port 4000 ...")
}
