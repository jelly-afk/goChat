package main

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID int
	mu     sync.RWMutex
}

type Hub struct {
	chatClients map[int]map[*Client]bool
	broadcast   chan []byte
	register    chan *Client
	unregister  chan *Client
	mu          sync.RWMutex
}

type Message struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	ChatID  int    `json:"chat_id"`
	UserID  int    `json:"user_id"`
}


func newHub() *Hub {
	return &Hub{
		chatClients: make(map[int]map[*Client]bool),
		broadcast:   make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.chatClients[client.userID]; !ok {
				h.chatClients[client.userID] = make(map[*Client]bool)
			}
			h.chatClients[client.userID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.chatClients[client.userID]; ok {
				delete(clients, client)
				if len(clients) == 0 {
					delete(h.chatClients, client.userID)
				}
			}
			h.mu.Unlock()
			close(client.send)
			client.conn.Close()

		case message := <-h.broadcast:
			var msg Message
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("error unmarshaling broadcast message: %v", err)
				continue
			}

			h.mu.RLock()
			clients, ok := h.chatClients[msg.ChatID]
			if !ok {
				h.mu.RUnlock()
				continue
			}
			for client, _ := range clients {
				if client.userID == msg.UserID {
					continue
				}
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(clients, client)
					client.conn.Close()
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("error unmarshaling message: %v", err)
			continue
		}

		msg.UserID = c.userID

		messageBytes, err := json.Marshal(msg)
		if err != nil {
			log.Printf("error marshaling message: %v", err)
			continue
		}

		c.hub.broadcast <- messageBytes
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}
