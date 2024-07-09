package router

import (
	"server/controllers"
	"github.com/gorilla/mux"
)

func Router() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/get", controllers.GetMyAllPosts).Methods("GET")
	router.HandleFunc("/api/v1/getimg", controllers.GenerateImageFromGetImg).Methods("POST","OPTIONS")
	router.HandleFunc("/api/v1/post", controllers.PostPhoto).Methods("POST","OPTIONS")
	return router
}
