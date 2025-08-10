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

type OpenedEventType struct {
	Type string `json:"type"`
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
	con.WriteJSON(openedEvent)
	go (func() {
		var writeError error = nil
		for writeError == nil {
			time.Sleep(time.Minute)
			writeError = con.WriteJSON(pingEvent)
			if writeError == nil {
				fmt.Println("Ping sent.")
			}
		}
	})()
	fmt.Println(con)
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
