package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type MarshalSyncMap struct {
	sync.Map
}

type AppState struct {
	rooms       MarshalSyncMap
	roomClients MarshalSyncMap
	userIds     MarshalSyncMap
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
	roomId       string
	userIdWasSet bool
	mutex        sync.Mutex
}

func (u *UserIdContainer) set(id string, roomId string) bool {
	defer u.mutex.Unlock()

	u.mutex.Lock()
	if u.userIdWasSet {
		return false
	}
	u.userIdWasSet = true
	u.userId = id
	u.roomId = roomId
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

type MessageEvent struct {
	Type string  `json:"type"`
	Data Message `json:"data"`
}

type SendEvent struct {
	Message string `json:"message"`
	Room    string `json:"room"`
}

func handleLoginEvent(con *websocket.Conn, byt []byte, userIdContainer *UserIdContainer, handle string) error {
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
	if !userIdContainer.set(userId, dat.RoomId) {
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
	addConnectionToRoom(con, dat.RoomId, handle)

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

	text := dat.Message
	maybeUuid, err := uuid.Parse(userId)
	if err != nil {
		return err
	}
	message := Message{
		Text:   text,
		Sender: maybeUuid,
		Time:   time.Now(),
		Id:     uuid.New(),
	}
	room := userIdContainer.roomId
	message.Sender = maybeUuid

	insertMessage(room, message)
	broadcastToRoom(room, message)

	return nil
}

func messageLoop(con *websocket.Conn, userId *UserIdContainer) {
	handle := uuid.NewString()

	for {
		_, byt, err := con.ReadMessage()
		if err != nil {
			if userId.userIdWasSet {
				removeConnectionFromRoom(userId.roomId, handle)
			}
			return
		}

		var dat SomeEvent
		if err := json.Unmarshal(byt, &dat); err != nil {
			continue
		}

		switch dat.Type {
		case "auth":
			handleLoginEvent(con, byt, userId, handle)
		case "send":
			handleSendEvent(con, byt, userId)
		}
	}
}

func addConnectionToRoom(c *websocket.Conn, room string, handle string) {
	roomMap_, ok := state.roomClients.Load(room)
	if !ok {
		roomMap_ = map[string]*websocket.Conn{}
	}
	roomMap := roomMap_.(map[string]*websocket.Conn)
	roomMap[handle] = c
	state.roomClients.Store(room, roomMap)
}

func broadcastToRoom(room string, message Message) {
	roomMap_, ok := state.roomClients.Load(room)
	if !ok {
		roomMap_ = map[string]*websocket.Conn{}
	}
	roomMap := roomMap_.(map[string]*websocket.Conn)

	for k := range roomMap {
		roomMap[k].WriteJSON(MessageEvent{
			Type: "newMessage",
			Data: message,
		})
	}

}

func removeConnectionFromRoom(room string, handle string) {
	roomMap_, ok := state.roomClients.Load(room)
	if !ok {
		roomMap_ = map[string]*websocket.Conn{}
	}
	roomMap := roomMap_.(map[string]*websocket.Conn)
	delete(roomMap, handle)
	state.roomClients.Store(room, roomMap)
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	con, _ := ws.Upgrade(w, r, nil)
	con.WriteJSON(openedEvent)

	userId := UserIdContainer{}

	go pingLoop(con)
	go messageLoop(con, &userId)
}

type PersistedData struct {
	RoomsStr string
	UsersStr string
}

func (f *MarshalSyncMap) UnmarshalRooms(data []byte) error {
	var tmpMap map[string][]Message
	if err := json.Unmarshal(data, &tmpMap); err != nil {
		return err
	}
	for key, value := range tmpMap {
		f.Store(key, value)
	}
	return nil
}
func (f *MarshalSyncMap) UnmarshalUserIds(data []byte) error {
	var tmpMap map[string]string
	if err := json.Unmarshal(data, &tmpMap); err != nil {
		return err
	}
	for key, value := range tmpMap {
		f.Store(key, value)
	}
	return nil
}

func (f *MarshalSyncMap) MarshalJSON() ([]byte, error) {
	tmpMap := make(map[string]any)
	f.Range(func(k, v any) bool {
		tmpMap[k.(string)] = v
		return true
	})
	return json.Marshal(tmpMap)
}

func tryRestore() {
	byt, err := os.ReadFile("data.json")
	if err != nil {
		return
	}

	var dat PersistedData
	if err := json.Unmarshal(byt, &dat); err != nil {
		return
	}

	state.rooms.UnmarshalRooms([]byte(dat.RoomsStr))
	state.userIds.UnmarshalUserIds([]byte(dat.UsersStr))
}

func main() {
	tryRestore()

	server := &http.Server{
		Addr: ":8081",
	}

	http.HandleFunc("/ws", websocketHandler)
	go func() { http.ListenAndServe(":8081", nil) }()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 3*time.Second)
	defer shutdownRelease()

	server.Shutdown(shutdownCtx)
	println("")

	RoomsStr, err := state.rooms.MarshalJSON()
	if err != nil {
		return
	}
	UsersStr, err := state.userIds.MarshalJSON()
	if err != nil {
		return
	}

	persisted := PersistedData{
		RoomsStr: string(RoomsStr),
		UsersStr: string(UsersStr),
	}
	persistedData, err := json.Marshal(persisted)
	if err != nil {
		return
	}

	os.WriteFile("data.json", persistedData, 0644)

	println("Goodbye.")
}
