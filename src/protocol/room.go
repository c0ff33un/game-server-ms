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
	//"github.com/coff33un/game-server-ms/src/common"
	//"go.mongodb.org/mongo-driver/bson"
	//"go.mongodb.org/mongo-driver/mongo/options"
	//"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Room struct {
	ID   string `json:"id"`
	hub  *Hub
	game *game.Game

	Ready   bool `json:"ready"`
	Setup   bool `json:"setup"`
	Running bool `json:"runnning"`
	Closing bool `json:"closing"`

	clients map[*Client]bool
	byid    map[string]*Client

	activity time.Time

	register   chan *Client
	unregister chan *Client
	broadcast  chan map[string]interface{}
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
		ID:      id,
		hub:     h,
		clients: make(map[*Client]bool),
		byid:    make(map[string]*Client),

		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan map[string]interface{}),
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

	for player := range r.clients {
		x, y := r.game.Begin.X, r.game.Begin.Y
		r.broadcast <- map[string]interface{}{
			"type": "move",
			"id":   player.ID,
			"x":    y,
			"y":    x,
		}
	}
	r.broadcast <- map[string]interface{}{
		"type": "start",
		"grid": r.game.Board.Grid,
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

func (r *Room) writeRoomInfo() {
	/*db := r.hub.db
	  ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	  err := db.Ping(ctx, readpref.Primary())
	  collection := client.Database("taurus").Collection("rooms")

	  ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
	  res, err := collection.InsertOne(ctx, bson.D())*/
}

func (r *Room) playerConnected(id string) bool {
	if c, ok := r.byid[id]; ok { // check registered
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
	return len(r.byid) < 3 && !r.Running && !r.Setup
}

func (r *Room) Close(quit chan struct{}, wg *sync.WaitGroup) {
	fmt.Println("Closing Room")
	r.Closing = true
	r.hub.unregister <- r
	for client := range r.clients {
		close(client.send)
		delete(r.clients, client)
	}
	wg.Wait()
	close(quit)
	log.Printf("Closed room %v", r.ID)
}

func (r *Room) cullInactive(minutes int, quit chan struct{}, wg *sync.WaitGroup) {
	log.Printf("Checking for room inactivity every: %v minutes\n", minutes)
	for {
		last := time.Now()
		time.Sleep(time.Duration(minutes) * time.Minute)
		log.Printf("Since Activity %v Since Last Check: %v\n", time.Since(r.activity), time.Since(last))
		if time.Since(r.activity) > time.Since(last) {
			r.Close(quit, wg)
			return
		}
	}
}

func (r *Room) Run() {
	defer func() {
		fmt.Println("Room Died")
	}()

	var wg sync.WaitGroup
	quit := make(chan struct{})
	go r.cullInactive(10, quit, &wg)
	for {
		fmt.Println("Room Activity...")
		r.activity = time.Now()
		select {
		case client := <-r.register:
			fmt.Println("Trying to register")
			r.clients[client] = true
			r.byid[client.ID] = client
			fmt.Println("Room Go: registered client", client.ID)
			wg.Add(2)
			go client.WritePump(&wg)
			go client.ReadPump(&wg)
		case client := <-r.unregister: // websocket closed
			fmt.Println("Unregistering Client:", client.ID)
			if _, ok := r.clients[client]; ok {
				close(client.send)
				delete(r.clients, client)
			}
		case message := <-r.broadcast:
			if message["type"] == "win" {
				message["handle"] = r.byid[message["id"].(string)].Handle
				r.StopGame()
			}
			for client := range r.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(r.clients, client)
				}
			}
		case <-quit:
			return
		}
	}
}
