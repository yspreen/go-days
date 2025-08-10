package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

var syncMap sync.Map

type Message struct {
	Sender uuid.UUID `json:"sender"`
	Text   string    `json:"text"`
	Time   time.Time `json:"time"`
	Id     uuid.UUID `json:"id"`
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
	http.ListenAndServe(":8081", nil)
}
