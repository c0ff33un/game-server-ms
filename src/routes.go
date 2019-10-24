package main

import (
  "net/http"
  //"github.com/coff33un/game-server-ms/src/protocol"
)

func routes() {
	http.HandleFunc("/room", handleRooms)
	http.HandleFunc("/room/setup", hub.SetupRoom)
	http.HandleFunc("/room/start", hub.StartRoom)
	http.HandleFunc("/ws", hub.ServeWs)
 }
