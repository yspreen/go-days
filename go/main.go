package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type AppState struct {
	rooms   sync.Map
	userIds sync.Map
}

var state = AppState{}

type Message struct {
	Sender uuid.UUID `json:"sender"`
	Text   string    `json:"text"`
	Time   time.Time `json:"time"`
	Id     uuid.UUID `json:"id"`
}

type OpenedEventType struct {
	Type string `json:"type"`
}

type AuthResponseEventType struct {
	Type   string `json:"type"`
	Secret string `json:"secret"`
	UserId string `json:"userId"`
}

var openedEvent = OpenedEventType{
	Type: "opened",
}

type PingEventType struct {
	Type string `json:"type"`
}

var pingEvent = PingEventType{
	Type: "ping",
}

func loadMessages(key string) ([]Message, bool) {
	val, ok := state.rooms.LoadOrStore(key, []Message{})
	return val.([]Message), ok
}

func insertMessage(key string, newMessage Message) {
	old, _ := loadMessages(key)
	state.rooms.Store(key, append(old, []Message{newMessage}...))
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}

var ws = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func pingLoop(con *websocket.Conn) {
	var writeError error = nil
	for writeError == nil {
		time.Sleep(time.Minute)
		writeError = con.WriteJSON(pingEvent)
		if writeError == nil {
			fmt.Println("Ping sent.")
		}
	}
}

type UserIdContainer struct {
	userId       string
	userIdWasSet bool
	mutex        sync.Mutex
}

func (u *UserIdContainer) set(id string) {
	defer u.mutex.Unlock()

	u.mutex.Lock()
	if u.userIdWasSet {
		return
	}
	u.userIdWasSet = true
	u.userId = id
}

type SomeMessage struct {
	Type string `json:"type"`
}

type LoginMessage struct {
	Secret string `json:"secret"`
}

func handleLoginMessage(con *websocket.Conn, byt []byte, userIdContainer *UserIdContainer) error {
	var dat LoginMessage
	if err := json.Unmarshal(byt, &dat); err != nil {
		return err
	}

	secret := dat.Secret
	userIdLoad, ok := state.userIds.Load(secret)
	if !ok {
		secret = uuid.NewString()
		userIdLoad = uuid.NewString()
		state.userIds.Store(secret, userIdLoad)
	}

	userId := userIdLoad.(string)
	userIdContainer.set(userId)

	con.WriteJSON(AuthResponseEventType{
		Type:   "authResponse",
		Secret: secret,
		UserId: userId,
	})

	return nil
}

func messageLoop(con *websocket.Conn, userId *UserIdContainer) {
	for {
		_, byt, err := con.ReadMessage()
		if err != nil {
			return
		}

		var dat SomeMessage
		if err := json.Unmarshal(byt, &dat); err != nil {
			continue
		}

		switch dat.Type {
		case "auth":
			handleLoginMessage(con, byt, userId)
		}
	}
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	con, _ := ws.Upgrade(w, r, nil)
	con.WriteJSON(openedEvent)

	userId := UserIdContainer{}

	go pingLoop(con)
	go messageLoop(con, &userId)
}

func main() {
	insertMessage("test", Message{
		Sender: uuid.New(),
		Id:     uuid.New(),
		Text:   "hi",
		Time:   time.Now(),
	})
	val, _ := loadMessages("test")
	json, _ := json.Marshal(val)
	fmt.Println("res: ", string(json))

	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/ws", websocketHandler)
	http.ListenAndServe(":8081", nil)
}
