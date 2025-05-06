package server

import (
	"encoding/json" 
	"log"         

	"tressette-game/internal/shared"
	"tressette-game/internal/protocol"

	"github.com/gorilla/websocket"
)

// Client represents a single WebSocket connection.
type Client struct {
	hub  			*Hub
	conn 			*websocket.Conn
	send 			chan []byte
	ID   			string // Unique identifier for the client/player
	Name 			string // Player's chosen name
	DesiredTeam 	shared.TeamEnum // Desired team for the player
}

// ReadPump handles incoming messages from the WebSocket connection.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if (err != nil) {
			log.Printf("Read error from client %s (%s): %v", c.ID, c.conn.RemoteAddr(), err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Unexpected close error: %v", err)
			}
			break // Exit loop on read error or connection close
		}

		var msg protocol.Message 
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("Error unmarshalling message from client %s: %v", c.ID, err)

			continue 
		}

		if (msg.Type != "ping") {
			log.Printf("Received message type '%s' from client %s (%s)", msg.Type, c.ID, c.Name)
		}
		c.hub.processMessage <- clientMessage{client: c, message: msg}
	}
}

// WritePump handles outgoing messages to the WebSocket connection.
func (c *Client) WritePump() {
	defer c.conn.Close()

	for message := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Write error to client %s (%s): %v", c.ID, c.Name, err) // Added logging
			break
		}
	}
}