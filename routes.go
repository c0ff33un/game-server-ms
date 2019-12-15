package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func routes() error {
	r := mux.NewRouter()
	s := r.PathPrefix(os.Getenv("API_PREFIX")).Subrouter()
	s.HandleFunc("/room", hub.CreateRoom)
	s.HandleFunc("/room/{roomid}", hub.GetRoom)
	s.HandleFunc("/room/setup/{roomid}", hub.SetupRoom)
	s.HandleFunc("/room/start/{roomid}", hub.StartRoom)
	s.HandleFunc("/room/setupready/{roomid}", hub.StartRoom)
	s.HandleFunc("/room/ready/{roomid}", hub.RoomReady)
	s.HandleFunc("/ws/{roomid}", hub.ServeWs)
	err := http.ListenAndServe(":8080", r)
	log.Println("Routes set up")
	return err
}
