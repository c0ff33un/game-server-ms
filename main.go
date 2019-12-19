// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/coff33un/game-server-ms/protocol"
)

//var addr = flag.String("addr", ":8080", "http service address")
var hub *protocol.Hub

func main() {
	if os.Getenv("NO_LOGS") != "" {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}
	hub = protocol.NewHub()
	go hub.Run()
	err := routes()
	if err != nil {
		log.Println(err)
		return
	}
}
