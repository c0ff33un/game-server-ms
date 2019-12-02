package protocol

import (
	"fmt"
	"log"
	//"context"
	"math/rand"
	"strconv"
	"sync"
	"time"

	//"github.com/segmentio/ksuid"
	"github.com/coff33un/game-server-ms/src/game"
)

type Room struct {
	ID          string `json:"_id"`
	Description string `json:"description"`
	Length      int    `json:"length"`
	Capacity    int    `json:"capacity"`
	Ready       bool   `json:"ready"`
	Setup       bool   `json:"setup"`
	Running     bool   `json:"runnning"`
	Closing     bool   `json:"closing"`
	Public      bool   `json:"public"`
	clients     map[*Client]bool
	byid        map[string]*Client
	activity    time.Time
	hub         *Hub
	game        *game.Game
	register    chan *Client
	unregister  chan *Client
	broadcast   chan map[string]interface{}
	unicast     chan map[string]interface{}
	quit        chan struct{}
	sync.WaitGroup
}

func genId() string {
	return strconv.Itoa(rand.Intn(10000))
}

func NewRoom(h *Hub) *Room {
	id := genId()
	for h.Byid[id] != nil {
		id = genId()
	}
	fmt.Println("New RoomID:", id)
	room := &Room{
		ID:         id,
		hub:        h,
		Length:     0,
		Capacity:   3,
		activity:   time.Now(),
		clients:    make(map[*Client]bool),
		byid:       make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan map[string]interface{}),
		unicast:    make(chan map[string]interface{}),
		quit:       make(chan struct{}),
	}
	room.hub.register <- room
	return room
}

func (r *Room) SetupGame(v game.SetupGameMessage) {
	fmt.Println("Setting up game.")

	var players []string
	for player := range r.clients {
		players = append(players, player.ID)
	}
	r.game = game.NewGame(v, r.broadcast, players)
	r.Setup = true
	r.Ready = true
	r.broadcast <- map[string]interface{}{
		"type": "setup",
	}
	fmt.Println("Room Setup")
}

func (r *Room) StartGame() {
	if !r.Ready || !r.Setup {
		fmt.Println("Room Not Setup")
		return
	}

	r.Running = true
	go r.game.Run()
}

func (r *Room) StopGame() {
	if !r.Running {
		fmt.Println("Cannot Stop A Room that is Not Running")
		return
	}
	fmt.Println("Try Stop Game")
	r.Running = false
	r.Ready = false
	r.Setup = false
	close(r.game.Quit)
	fmt.Println("Gracefully Stopped Game")
}

func (r *Room) Close() {
	fmt.Println("Closing Room")
	if r.Closing {
		return
	}
	r.Closing = true
	if r.Running {
		r.StopGame()
	}
	r.hub.unregister <- r
	for client := range r.clients {
		close(client.send)
		delete(r.clients, client)
	}
	r.Wait() // Anonymous WaitGroup
	close(r.quit)
	log.Printf("Closed room %v", r.ID)
}

func (r *Room) writeRoomInfo() {
	/*db := r.hub.db
	  ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	  err := db.Ping(ctx, readpref.Primary())
	  collection := client.Database("taurus").Collection("rooms")

	  ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
	  res, err := collection.InsertOne(ctx, bson.D())*/
}

func (r *Room) playerConnected(id string) bool {
	if c, ok := r.byid[id]; ok { // check Registered
		_, ok := r.clients[c] // check WS Connection
		return ok
	}
	return false
}

func (r *Room) OkToConnectPlayer(id string) bool {
	if r.Closing {
		return false
	}
	if r.playerConnected(id) {
		fmt.Println("User already Connected to Room")
		return false // there can't be two connections of the same user
	}
	if _, ok := r.byid[id]; ok {
		fmt.Println("User is registrered and Disconnected")
		return true
	}
	fmt.Println("Room len", len(r.byid))
	return r.Length < r.Capacity && !r.Running && !r.Setup
}

func (r *Room) send(client *Client, message map[string]interface{}) {
	select {
	case client.send <- message:
	default:
		close(client.send)
		delete(r.clients, client)
	}
}

func (r *Room) cullInactive(minutes int) {
	log.Printf("Checking for room inactivity every: %v minutes\n", minutes)
	duration := time.Duration(minutes) * time.Minute
	for {
		select {
		case <-r.quit:
			return
		case <-time.After(duration):
			if time.Since(r.activity) > duration {
				r.Close()
				return
			}
		}
	}
}

func (r *Room) Run() {
	defer func() {
		fmt.Println("Room Died")
	}()
	go r.cullInactive(10)
	for {
		select {
		case <-r.quit:
			return
		case client := <-r.register:
			fmt.Println("Trying to register")
			r.clients[client] = true
			r.byid[client.ID] = client
			fmt.Println("Room Go: registered client", client.ID)
			r.Length += 1
			r.Add(2)
			go client.WritePump()
			go client.ReadPump()
			if r.Running {
				for _, player := range r.game.Players {
					fmt.Println(player.MoveMessage())
					r.send(client, player.MoveMessage())
				}
				r.send(client, r.game.BoardMessage())
			}
		case client := <-r.unregister: // websocket closed
			fmt.Println("Unregistering Client:", client.ID)
			close(client.send)
			delete(r.clients, client)
			if !r.Running {
				// Allow users to reconnect even if room is full.
				// Create Space in Room
				delete(r.byid, client.ID)
				r.Length -= 1
				if r.Length == 0 {
					r.Close()
				}
			}
		case message := <-r.unicast:
			id := message["id"].(string)
			client := r.byid[id]
			r.send(client, message)
		case message := <-r.broadcast:
			if message["type"] == "win" {
				message["handle"] = r.byid[message["id"].(string)].Handle
				r.StopGame()
			}
			for client := range r.clients {
				r.send(client, message)
			}
		}
	}
}
