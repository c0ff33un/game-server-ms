// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protocol

import (
	//"bytes"
	"log"
	"time"
	"fmt"
	"io"
	"encoding/json"

	"github.com/gorilla/websocket"
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

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Client is a mIDdleman between the websocket connection and the hub.
type Client struct {
	room *Room

	// user ID
	ID string

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan interface{}
}

func printJSON(m map[string]interface{}) {
  fmt.Println("JSON:")
  for k, v := range m {
    switch vv := v.(type) {
      case string:
          fmt.Println(k, "is string", vv)
      case float64:
          fmt.Println(k, "is float64", vv)
      case int:
          fmt.Println(k, "is int", vv)
      case []interface{}:
          fmt.Println(k, "is an array:")
          for i, u := range vv {
              fmt.Println(i, u)
          }
      default:
          fmt.Println(k, "is of a type I don't know how to handle")
    }
  }
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) ReadPump() {
	defer func() {
		c.room.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
	  var v interface{}
		err := c.conn.ReadJSON(&v)
		if err != nil {
		  fmt.Println("Error", err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		m := v.(map[string]interface{})
    printJSON(m)
    // Server accepts WebSocket connection but needs ID and/or authentication to continue with requests
    // Should be subsequent to websocket connection
    if c.ID == "" {
      if m["type"] == "connect" && m["ID"] != "" {
        c.ID = m["id"].(string)
        c.room.broadcast <- interface{}(m)
      }
    } else {
      m["id"] = c.ID // Backend keeps the Clients IDs not Frontend
      switch m["type"] {
        case "message":
          fmt.Println("Broadcasting Message:")
          c.room.broadcast <- interface{}(m)
        case "move":
          c.room.game.Update <- interface{}(m)
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
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			printJSON(message.(map[string]interface{}))
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage) // Uses this instead of c.coon.WriteJSON to reuse NextWriter
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

