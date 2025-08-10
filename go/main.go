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

var syncMap sync.Map

type Message struct {
	Sender uuid.UUID `json:"sender"`
	Text   string    `json:"text"`
	Time   time.Time `json:"time"`
	Id     uuid.UUID `json:"id"`
}

type OpenedEvent_ struct {
	Type string `json:"type"`
}

var OpenedEvent = OpenedEvent_{
	Type: "opened",
}

func loadMessages(key string) ([]Message, bool) {
	val, ok := syncMap.LoadOrStore(key, []Message{})
	return val.([]Message), ok
}

func insertMessage(key string, newMessage Message) {
	old, _ := loadMessages(key)
	syncMap.Store(key, append(old, []Message{newMessage}...))
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}

var ws = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	con, _ := ws.Upgrade(w, r, nil)
	con.WriteJSON(OpenedEvent)
	fmt.Print(con)
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
	fmt.Print("res: ", string(json))

	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/ws", websocketHandler)
	http.ListenAndServe(":8081", nil)
}
