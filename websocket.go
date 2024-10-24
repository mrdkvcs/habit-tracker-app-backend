package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	connections = make(map[string]*websocket.Conn)
	broadcast   = make(chan string)
	mux         sync.Mutex
)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	userId := r.URL.Query().Get("userid")
	if err != nil {
		respondWithJson(w, 500, "Error upgrading ws connection")
		return
	}
	defer conn.Close()
	mux.Lock()
	connections[userId] = conn
	mux.Unlock()
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			mux.Lock()
			delete(connections, userId)
			mux.Unlock()
			break
		}
	}
}

func handleMessages() {
	for {
		recipientId := <-broadcast
		mux.Lock()
		conn, ok := connections[recipientId]
		mux.Unlock()
		if ok {
			err := conn.WriteMessage(websocket.TextMessage, []byte("Invite received"))
			if err != nil {
				log.Println("Error writing message to connection")
				mux.Lock()
				delete(connections, recipientId)
				mux.Unlock()
			}
		} else {
			log.Println("User not found")
		}
	}
}
