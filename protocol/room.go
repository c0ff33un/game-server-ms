package protocol

import (
	"context"
	"errors"
	"github.com/machinebox/graphql"
	"io"
	"log"
	// "math/rand"
	"os"
	// "strconv"
	"sync"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/coff33un/game-server-ms/game"
)

type Room struct {
	ID          string `json:"id" bson:"_id"`
	Description string `json:"description"`
	Capacity    int    `json:"capacity"`
	Length      int    `json:"length"`
	Ready       bool   `json:"ready"`
	Setup       bool   `json:"setup"`
	Running     bool   `json:"runnning"`
	Closing     bool   `json:"closing"`
	Public      bool   `json:"public"`

	clients        map[*Client]bool
	byid           map[string]*Client
	activity       time.Time
	hub            *Hub
	game           *game.GameRunner
	register       chan *Client
	unregister     chan *Client
	broadcast      chan interface{}
	quit           chan struct{}
	sync.WaitGroup `bson:"-"`
}

func genId() string {
	// return strconv.Itoa(rand.Intn(10000))
	return ksuid.New().String()
}

func NewRoom(h *Hub) *Room {
	id := genId()
	// for h.Byid[id] != nil {
	// 	id = genId()
	// }
	log.Println("New RoomID:", id)
	room := &Room{
		ID:         id,
		hub:        h,
		Capacity:   3,
		activity:   time.Now(),
		clients:    make(map[*Client]bool),
		byid:       make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan interface{}),
		quit:       make(chan struct{}),
	}
	room.hub.register <- room
	return room
}

type IdResponse struct {
	Id string
}

func (r *Room) writeMatch(message game.WinMessage) error {
	log.Printf("Message: %v\n", message)
	client := graphql.NewClient(os.Getenv("GRAPHQL_URL"))
	req := graphql.NewRequest(`
		mutation ($winner: String, $players: [String], $time: Int){
			newMatch(winner: $winner, players: $players, resolveTime: $time){ 
				id
			}
		}
	`)
	players := ""
	first := true
	for id := range r.byid {
		if first {
			first = false
			players = players + id
			continue
		}
		players = players + "," + id
	}
	req.Var("players", players)
	req.Var("time", message.ResolveTime)
	req.Var("winner", message.Id)
	//req.Header.Set("Accept", "application/json")
	ctx := context.TODO()
	respData := IdResponse{}
	err := client.Run(ctx, req, &respData)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (r *Room) SetupGame(body io.ReadCloser) error {
	log.Println("Setting up game.")

	var players []string
	for player := range r.clients {
		players = append(players, player.Id)
	}
	game, err := game.ClassicGameRunner(body, r.broadcast, players)
	if err != nil {
		return err
	}
	r.game = game
	r.Setup, r.Ready = true, true
	r.broadcast <- map[string]interface{}{
		"type": "setup",
	}
	log.Println("Room Succesfully Setup")
	return nil
}

func (r *Room) StartGame() error {
	if !r.Ready || !r.Setup {
		log.Println("Room Not Setup")
		return errors.New("Room Not Setup")
	}

	r.Running = true
	go r.game.Run()
	return nil
}

func (r *Room) StopGame() {
	if !r.Running {
		log.Println("Cannot Stop A Room that is Not Running")
		return
	}
	log.Println("Try Stop Game")
	r.Running = false
	r.Ready = false
	r.Setup = false
	close(r.game.Quit)
	log.Println("Gracefully Stopped Game")
}

func (r *Room) Close() {
	log.Println("Closing Room")
	if r.Closing {
		return
	}
	r.Closing = true
	if r.Running {
		r.StopGame()
	}
	r.hub.unregister <- r
	for client := range r.clients {
		client.conn.Close()
	}
	r.Wait() // Anonymous WaitGroup
	close(r.quit)
	log.Printf("Closed room %v", r.ID)
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
		log.Println("User already Connected to Room")
		return false // there can't be two connections of the same user
	}
	if _, ok := r.byid[id]; ok {
		log.Println("User is registrered and Disconnected")
		return true
	}
	return r.Length < r.Capacity && !r.Running && !r.Setup
}

func (r *Room) send(client *Client, message interface{}) {
	select {
	case client.send <- message:
	default:
		close(client.send)
		delete(r.clients, client)
	}
}

func (r *Room) sendBroadCast(message interface{}) {
	for client := range r.clients {
		r.send(client, message)
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
				log.Println("Closing Room due to inactivity")
				r.Close()
				return
			}
		}
	}
}

func (r *Room) Run() {
	defer func() {
		log.Println("Room Died")
	}()
	go r.cullInactive(15)
	for {
		log.Println("Room Run here..")
		r.activity = time.Now()
		select {
		case <-r.quit:
			return
		case client := <-r.register:
			log.Println("Trying to register")
			r.clients[client] = true
			r.byid[client.Id] = client
			r.Length = len(r.byid)
			log.Println("Room Go: registered client", client.Id)
			r.Add(2)
			go client.WritePump()
			go client.ReadPump()
			if r.Running {
				for _, f := range r.game.GetStatus() {
					r.send(client, f)
				}
			}
		case client := <-r.unregister: // websocket closed
			log.Println("Unregistering Client:", client.Id)
			delete(r.clients, client)
			if client.Leave {
				log.Println("Client Leaving")
				delete(r.byid, client.Id)
				r.Length = len(r.byid)
				r.sendBroadCast(client.getLeaveMessage(r.Length))
			}
			if len(r.byid) == 0 {
				log.Println("Closing room due to players disconnection")
				r.Close()
			}
		case message := <-r.broadcast:
			if message, ok := message.(game.WinMessage); ok {
				message.Handle = r.byid[message.Id].Handle
				r.writeMatch(message)
				r.sendBroadCast(message)
				r.StopGame()
				continue
			}
			r.sendBroadCast(message)
		}
	}
}
