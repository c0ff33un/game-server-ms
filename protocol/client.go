// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protocol

import (
	//"bytes"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/coff33un/game-server-ms/game"
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

// Client is a mIDdleman between the websocket connection and the hub.
type Client struct {
	room *Room

	// user ID
	Id string

	Handle string

	Leave bool

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan interface{}
}

type message struct {
	Type string `json:"type"`
}

type textMessage struct {
	message
	Text string `json:"text"`
}

type connectMessage struct {
	message
	Handle string `json:"handle"`
	Length int    `json:"length"`
}

func (c *Client) getLeaveMessage(length int) interface{} {
	return leaveMessage{
		message: message{Type: "leave"},
		Handle:  c.Handle,
		Length:  length,
	}
}

type leaveMessage struct {
	message
	Handle string `json:"handle"`
	Length int    `json:"length"`
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) ReadPump() {
	log.Println("ReadPump Started")
	defer func() {
		log.Println("ReadPump Ended")
		c.room.unregister <- c
		c.conn.Close()
		c.room.Done() // Anonymous WaitGroup
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("Error", err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v\n", err)
			}
			break
		}

		var mp map[string]interface{}
		err = json.Unmarshal(message, &mp)
		if err != nil {
			continue
		}

		Type := mp["type"].(string)
		switch Type {
		case "message":
			m := textMessage{}
			json.Unmarshal(message, &m)
			c.room.broadcast <- m
		case "connect":
			m := connectMessage{}
			json.Unmarshal(message, &m)
			m.Handle = c.Handle
			m.Length = c.room.Length
			c.room.broadcast <- m
		case "move":
			if c.room.Running {
				m := game.ClientMoveMessage{}
				json.Unmarshal(message, &m)
				m.Id = c.Id
				c.room.game.Update <- m
			}
		case "leave":
			c.Leave = true
			c.room.unregister <- c
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
	log.Println("WritePump Started")
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		log.Println("WritePump Ended")
		ticker.Stop()
		c.conn.Close()
		c.room.Done() // Anonymous WaitGroup
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				log.Println("The hub closed the channel")
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
