package main

import (
  "os"
  "fmt"
  "net/http"
  //"github.com/coff33un/game-server-ms/src/protocol"

  "github.com/gorilla/mux"
)

func routes() error {
  r := mux.NewRouter()
  s := r.PathPrefix(os.Getenv("API_PREFIX")).Subrouter()
	s.HandleFunc("/room", createRoom)
	s.HandleFunc("/room/{roomid}", getRoom)
	s.HandleFunc("/room/setup/{roomid}", hub.SetupRoom)
	s.HandleFunc("/room/start/{roomid}", hub.StartRoom)
	s.HandleFunc("/room/setupready/{roomid}", hub.StartRoom)
  s.HandleFunc("/room/ready/{roomid}", hub.RoomReady)
	s.HandleFunc("/ws/{roomid}", hub.ServeWs)
  err := http.ListenAndServe(":8080", r)
  fmt.Println("Routes set up")
  return err
 }
