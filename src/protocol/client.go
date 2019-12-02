// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protocol

import (
	//"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	//"github.com/coff33un/game-server-ms/src/common"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// Client is a mIDdleman between the websocket connection and the hub.
type Client struct {
	room *Room

	// user ID
	ID string

	Handle string

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan map[string]interface{}
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) ReadPump(wg *sync.WaitGroup) {
	fmt.Println("ReadPump Started")
	defer func() {
		fmt.Println("ReadPump Ended")
		c.room.unregister <- c
		c.conn.Close()
		wg.Done()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		var m map[string]interface{}
		err := c.conn.ReadJSON(&m)
		if err != nil {
			fmt.Println("Error", err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		//common.PrintJSON(m)
		m["id"] = c.ID // Backend keeps the Clients IDs not Frontend
		switch m["type"] {
		case "win":
			c.room.StopGame()
			c.room.broadcast <- m
		case "connect":
			m["handle"] = c.Handle
			c.room.broadcast <- m
		case "message":
			c.room.broadcast <- m
		case "move":
			if c.room.Running {
				c.room.game.Update <- m
			}
		}
	}
}

func writeJSON(w io.WriteCloser, v interface{}) error {
	err := json.NewEncoder(w).Encode(v)
	return err
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) WritePump(wg *sync.WaitGroup) {
	fmt.Println("WritePump Started")
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		fmt.Println("WritePump Ended")
		ticker.Stop()
		c.conn.Close()
		wg.Done()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				fmt.Println("The hub closed the channel")
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			//common.PrintJSON(message.(map[string]interface{}))

			w, err := c.conn.NextWriter(websocket.TextMessage) // Uses this instead of c.conn.WriteJSON to reuse Writer
			if err != nil {
				return
			}
			writeJSON(w, message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				writeJSON(w, <-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
