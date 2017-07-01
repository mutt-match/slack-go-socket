package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/googollee/go-socket.io"
)

// Message : message construct to keep track of posts in rooms
type Message struct {
	User         User      `json:"user"`
	RoomMessages []Message `json:"roomMessages"`
	Sender       string    `json:"sender"`
	CreatedAt    time.Time `json:"created_at"`
	Room         string    `json:"room"`
	Text         string    `json:"message"`
	Participants []User    `json:"participants"`
}

// User : users connected the the app through socket
type User struct {
	Name   string `json:"name"`
	ID     int32  `json:"id"`
	Socket string `json:"socket"`
	Room   string `json:"room"`
}

func remove(a []User, i int) []User {
	log.Println("index", i, "users", a)
	a[i] = a[len(a)-1]
	return a[:len(a)-1]
}

func main() {

	var participants = []User{}
	messages := make(map[string][]Message)
	// var nameCounter = 1

	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	server.On("connection", func(socket socketio.Socket) {
		log.Println("on connection to socket", socket.Id())
		// socket.Join(room)
		// testUser := User{
		// 	Name:   "Max",
		// 	ID:     "1234",
		// 	Socket: "12edad23rr232345",
		// }

		// testMessage := Message{
		// 	User:      testUser,
		// 	Sender:    "system",
		// 	CreatedAt: time.Now().Local(),
		// 	Text:      "joined room",
		// 	Room:      room,
		// }

		// testMessage.MarshalJSON()
		// testMessage = json.Marshal(testMessage)

		// server.BroadcastTo(room, "join", testMessage)

		socket.On("add:user", func(data string) {
			var newUser = User{}
			pointer := &newUser
			bytes := []byte(data)
			if err := json.Unmarshal(bytes, pointer); err != nil {
				log.Fatal(err)
			}

			pointer.Socket = socket.Id()

			participants = append(participants, newUser)

			log.Println("participants", participants)
			newMessage := Message{
				User:         newUser,
				RoomMessages: messages[newUser.Room],
				Sender:       "system",
				CreatedAt:    time.Now().Local(),
				Room:         newUser.Room,
				Text:         "",
				Participants: participants,
			}

			if _, err := json.Marshal(newMessage); err != nil {
				panic(err)
			} else {
				server.BroadcastTo(newUser.Room, "new:user", newMessage)
			}
		})

		socket.On("join", func(data string) {

			log.Println("join", data)
			var parsedData = Message{}
			pointer := &parsedData
			bytes := []byte(data)
			if err := json.Unmarshal(bytes, pointer); err != nil {
				log.Fatal(err)
			}

			if _, new := messages[parsedData.Room]; !new {
				messages[parsedData.Room] = []Message{}
			}
			var newMessage = Message{
				User:         parsedData.User,
				Room:         parsedData.Room,
				RoomMessages: messages[parsedData.Room],
			}
			log.Println("join room", parsedData.Room)
			socket.Join(parsedData.Room)
			server.BroadcastTo(parsedData.Room, "new:message", newMessage)
		})

		socket.On("add:message", func(data string) {

			log.Println("join", data)
			var parsedData = Message{}
			pointer := &parsedData
			bytes := []byte(data)
			if err := json.Unmarshal(bytes, pointer); err != nil {
				log.Fatal(err)
			}
			log.Println("new message", parsedData)
			messages[parsedData.Room] = append(messages[parsedData.Room], parsedData)
			log.Println("added message", parsedData)
			server.BroadcastTo(parsedData.Room, "new:message", parsedData)
		})

		socket.On("disconnection", func() {
			//remove participant
			log.Println("on disconnect")
			for idx := range participants {
				log.Println("index", idx)
				if socket.Id() == participants[idx].Socket {
					participants = remove(participants, idx)
					break
				}
			}
			user := User{Socket: socket.Id()}
			server.BroadcastTo("Lobby", "disconnect:user", user)
		})
	})

	server.On("disconnect:user", func(socket socketio.Socket, err error) {
		log.Println("disconnet:")
	})

	server.On("error", func(socket socketio.Socket, err error) {
		log.Println("error:", err)
	})

	// http.Handle("/socket.io/", server)
	http.HandleFunc("/socket.io/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "http://localhost:3456")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		server.ServeHTTP(w, r)
	})
	http.Handle("/", http.FileServer(http.Dir("./asset")))
	log.Println("Serving at http://localhost:5000")
	log.Fatal(http.ListenAndServe(":5000", nil))
}
