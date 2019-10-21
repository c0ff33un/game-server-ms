// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
  "fmt"
	"flag"
	"log"
	"net/http"
	"encoding/json"

	"github.com/coff33un/game-server-ms/src/protocol"
)

var addr = flag.String("addr", ":8080", "http service address")
var hub *protocol.Hub

// ruta para crear room, retorna el id (PseudoRandom) del room creado el room funciona como go rutina

func enableCors(w *http.ResponseWriter) {
  (*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func handleRooms(w http.ResponseWriter, r *http.Request) {
  switch (r.Method) {
  case http.MethodGet:
    enableCors(&w)
    err := r.ParseForm()
    if err != nil {
      log.Println(err)
    }
    id := r.Form.Get("id")
    w.Header().Set("Content-Type", "application/json")
    if room, ok := hub.Byid[id]; ok {
      json.NewEncoder(w).Encode(room)
    } else {
      http.NotFound(w, r)
    }
    break;
  case http.MethodPost:
    enableCors(&w)
    room := protocol.NewRoom(hub)
    js, err := json.Marshal(map[string]string{"id": room.ID})
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
    fmt.Println("Created room")
    go room.Run()
    fmt.Println("room Run running")
    w.Header().Set("Content-Type", "application/json")
    w.Write(js)
    break;
  }
}

// ruta para unirse al room creado (websocket)

func main() {
	flag.Parse()
	hub = protocol.NewHub()
	go hub.Run()
	routes()
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
