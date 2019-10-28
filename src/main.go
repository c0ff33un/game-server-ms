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
	"github.com/coff33un/game-server-ms/src/common"
	"github.com/gorilla/mux"
)

//var addr = flag.String("addr", ":8080", "http service address")
var hub *protocol.Hub

// ruta para crear room, retorna el id (PseudoRandom) del room creado el room funciona como go rutina

func enableCors(w *http.ResponseWriter) {
  (*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func getRoom(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  fmt.Println("here")
  common.DisableCors(&w)
  switch (r.Method) {
  case http.MethodGet:
    err := r.ParseForm()
    if err != nil {
      log.Println(err)
      return
    }
    id := vars["roomid"]
    w.Header().Set("Content-Type", "application/json")
    if room, ok := hub.Byid[id]; ok {
      json.NewEncoder(w).Encode(room)
    } else {
      http.NotFound(w, r)
    }
  }
}

func createRoom(w http.ResponseWriter, r *http.Request) {
  common.DisableCors(&w)
  switch (r.Method) {
  case http.MethodPost:
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
  }
}

// ruta para unirse al room creado (websocket)

func main() {
	flag.Parse()
	hub = protocol.NewHub()
	go hub.Run()
  err := routes()
  if err != nil {
    return
  }
}
