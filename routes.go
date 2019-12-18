package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return ":8080"
	}
	return ":" + port
}

func routes() error {
	r := mux.NewRouter()
	s := r.PathPrefix(os.Getenv("API_PREFIX")).Subrouter()
	s.HandleFunc("/room", hub.CreateRoom)
	s.HandleFunc("/room/{roomid}", hub.GetRoom)
	s.HandleFunc("/room/{roomid}/setup", hub.SetupRoom)
	s.HandleFunc("/room/{roomid}/start", hub.StartRoom)
	s.HandleFunc("/room/{roomid}/setupready", hub.StartRoom)
	s.HandleFunc("/room/{roomid}/ready", hub.RoomReady)
	s.HandleFunc("/ws/{roomid}", hub.ServeWs)
	err := http.ListenAndServe(getPort(), r)
	log.Println("Routes set up")
	return err
}
