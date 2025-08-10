package main

import (
	"encoding/json"
	"errors"
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

func (u *UserIdContainer) set(id string) bool {
	defer u.mutex.Unlock()

	u.mutex.Lock()
	if u.userIdWasSet {
		return false
	}
	u.userIdWasSet = true
	u.userId = id
	return true
}

type SomeEvent struct {
	Type string `json:"type"`
}

type AuthEvent struct {
	Secret string `json:"secret"`
	RoomId string `json:"roomId"`
}

type MessagesEvent struct {
	Type string    `json:"type"`
	Data []Message `json:"data"`
}

type SendEvent struct {
	Data Message `json:"data"`
	Room string  `json:"room"`
}

func handleLoginEvent(con *websocket.Conn, byt []byte, userIdContainer *UserIdContainer) error {
	var dat AuthEvent
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
	if !userIdContainer.set(userId) {
		return errors.New("set id not ok")
	}

	con.WriteJSON(AuthResponseEventType{
		Type:   "authResponse",
		Secret: secret,
		UserId: userId,
	})

	msg, ok := state.rooms.Load(dat.RoomId)
	if !ok {
		msg = []Message{}
	}
	con.WriteJSON(MessagesEvent{
		Type: "newMessages",
		Data: msg.([]Message),
	})

	return nil
}

func handleSendEvent(con *websocket.Conn, byt []byte, userIdContainer *UserIdContainer) error {
	var dat SendEvent
	if err := json.Unmarshal(byt, &dat); err != nil {
		return err
	}

	if !userIdContainer.userIdWasSet {
		return errors.New("not logged in")
	}

	userId := userIdContainer.userId

	message := dat.Data
	room := dat.Room
	maybeUuid, err := uuid.Parse(userId)
	if err != nil {
		return err
	}
	message.Sender = maybeUuid

	insertMessage(room, message)

	return nil
}

func messageLoop(con *websocket.Conn, userId *UserIdContainer) {
	for {
		_, byt, err := con.ReadMessage()
		if err != nil {
			return
		}

		var dat SomeEvent
		if err := json.Unmarshal(byt, &dat); err != nil {
			continue
		}

		switch dat.Type {
		case "auth":
			handleLoginEvent(con, byt, userId)
		case "send":
			handleSendEvent(con, byt, userId)
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
	http.HandleFunc("/ws", websocketHandler)
	http.ListenAndServe(":8081", nil)
}
