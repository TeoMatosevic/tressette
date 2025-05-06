package server

import (
	"encoding/json"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"
	"tressette-game/internal/game"
	"tressette-game/internal/protocol"
	"tressette-game/internal/shared"

	"github.com/google/uuid"
)

// clientMessage is a helper struct to pass messages along with the client reference.
type clientMessage struct {
	client  	*Client
	message 	protocol.Message
}

const gameCodeLength = 5 // Length of the unique game code

// Hub manages active WebSocket connections, lobbies, and game rooms.
type Hub struct {
	clients      	map[*Client]bool
	lobbies      	map[string][]*Client    // Map game code to list of clients in the lobby
	games        	map[string]*game.Game 	// Map game code to game instance
	clientToGame 	map[*Client]string   	// Map client to game code (lobby or active game)
	processMessage 	chan clientMessage
	register       	chan *Client
	unregister     	chan *Client
	clientMu     	sync.RWMutex 
	lobbyMu      	sync.RWMutex 
	gameMu       	sync.RWMutex 
	rng          	*rand.Rand  
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	// Seed the random number generator
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	return &Hub{
		clients:        make(map[*Client]bool),
		lobbies:        make(map[string][]*Client),
		games:          make(map[string]*game.Game),
		clientToGame:   make(map[*Client]string),
		processMessage: make(chan clientMessage),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		rng:            rng,
	}
}

// generateGameCode creates a unique alphanumeric game code.
func (h *Hub) generateGameCode() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for {
		var sb strings.Builder
		for i := 0; i < gameCodeLength; i++ {
			sb.WriteByte(letters[h.rng.Intn(len(letters))])
		}
		code := sb.String()

		// Check for collisions (read lock is sufficient)
		h.lobbyMu.RLock()
		_, lobbyExists := h.lobbies[code]
		h.lobbyMu.RUnlock()

		h.gameMu.RLock()
		_, gameExists := h.games[code]
		h.gameMu.RUnlock()

		if !lobbyExists && !gameExists {
			return code
		}
		log.Printf("Generated game code %s collided, retrying...", code)
	}
}

// Run starts the Hub's main loop.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			client.ID = uuid.NewString() // Assign a unique ID upon registration
			log.Printf("Client %s (%s) connected", client.ID, client.conn.RemoteAddr())
			h.clientMu.Lock()
			h.clients[client] = true
			h.clientMu.Unlock()

		case client := <-h.unregister:
			h.clientMu.Lock()
			gameCode, inGameOrLobby := h.clientToGame[client]
			_, clientExists := h.clients[client]

			if clientExists {
				delete(h.clients, client)
				delete(h.clientToGame, client) // Remove client from clientToGame mapping
				close(client.send)
				log.Printf("Client %s (%s) disconnected", client.ID, client.Name)
			}
			h.clientMu.Unlock() // Unlock clientMu before potentially locking others

			if inGameOrLobby {
				// Check if it was a lobby first
				h.lobbyMu.Lock()
				lobby, lobbyExists := h.lobbies[gameCode]
				if lobbyExists {
					// Remove client from lobby
					newLobby := []*Client{}
					for _, c := range lobby {
						if c != client {
							newLobby = append(newLobby, c)
						}
					}
					if len(newLobby) > 0 {
						h.lobbies[gameCode] = newLobby
						log.Printf("Client %s removed from lobby %s.", client.ID, gameCode)
						// Broadcast updated lobby state
						h.broadcastLobbyUpdate(gameCode, newLobby)
					} else {
						// Last player left, delete lobby
						delete(h.lobbies, gameCode)
						log.Printf("Client %s left lobby %s. Lobby deleted.", client.ID, gameCode)
					}
					h.lobbyMu.Unlock() // Unlock lobbyMu after lobby modification
				} else {
					h.lobbyMu.Unlock() // Unlock lobbyMu if not found in lobbies

					// If not in a lobby, check if they were in an active game
					h.gameMu.RLock()
					gameInstance, gameExists := h.games[gameCode]
					h.gameMu.RUnlock()

					if gameExists {
						log.Printf("Client %s was in game %s. Notifying game.", client.ID, gameCode)
						// Notify the game instance about the disconnect
						go gameInstance.HandlePlayerDisconnect(client.ID) // Run in goroutine to avoid blocking hub
						// TODO: Game cleanup logic might be needed here or triggered by the game instance setting its state.
					} else {
						log.Printf("Client %s disconnected but was mapped to non-existent game/lobby code %s", client.ID, gameCode)
					}
				}
			} else if clientExists {
				// Client existed but wasn't in a game or lobby (e.g., disconnected before joining/creating)
				log.Printf("Client %s disconnected before joining/creating a game.", client.ID)
			}


		case clientMsg := <-h.processMessage:
			// Process the message based on its type
			h.handleMessage(clientMsg.client, clientMsg.message)
		}
	}
}

// handleMessage processes a message received from a client.
func (h *Hub) handleMessage(client *Client, msg protocol.Message) {

	switch msg.Type {
	case "create_game":
		h.handleCreateGame(client, msg)
	case "join_game":
		h.handleJoinGame(client, msg)
	case "play_card", "declare":
		h.handleGameAction(client, msg)
	case "ping":
		pongMsg, _ := protocol.NewMessage("pong", nil) 
		client.send <- pongMsg
	default:
		log.Printf("Received unknown message type '%s' from client %s (%s)", msg.Type, client.ID, client.Name)
		h.sendErrorToClient(client, "Unknown message type.")
	}
}

// handleCreateGame handles a request to create a new game lobby.
func (h *Hub) handleCreateGame(client *Client, msg protocol.Message) {
	h.clientMu.RLock()
	_, alreadyInGame := h.clientToGame[client]
	h.clientMu.RUnlock()
	if alreadyInGame {
		log.Printf("Client %s tried to create game but is already associated with one.", client.ID)
		h.sendErrorToClient(client, "Already in a game or lobby.")
		return
	}

	var payload protocol.CreateGamePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		log.Printf("Error unmarshalling create_game payload from client %s: %v", client.ID, err)
		h.sendErrorToClient(client, "Invalid create_game message format.")
		return
	}
	if payload.Name == "" {
		log.Printf("Client %s tried to create game with an empty name.", client.ID)
		h.sendErrorToClient(client, "Name cannot be empty.")
		return
	}

	// Generate unique game code
	gameCode := h.generateGameCode()

	// Update client state and create lobby
	h.clientMu.Lock()
	client.Name = payload.Name
	client.DesiredTeam = payload.DesiredTeam // Set desired team
	h.clientToGame[client] = gameCode
	h.clientMu.Unlock()

	h.lobbyMu.Lock()
	h.lobbies[gameCode] = []*Client{client}
	h.lobbyMu.Unlock()

	log.Printf("Client %s (%s) created lobby %s", client.ID, client.Name, gameCode)

	// Send confirmation and lobby state back to creator
	createdPayload := protocol.GameCreatedPayload{GameCode: gameCode}
	createdMsg, _ := protocol.NewMessage("game_created", createdPayload)
	h.sendMessageToClient(client.ID, createdMsg)

	h.broadcastLobbyUpdate(gameCode, []*Client{client}) // Send initial lobby state
}

// handleJoinGame handles a request to join an existing game lobby.
func (h *Hub) handleJoinGame(client *Client, msg protocol.Message) {
	h.clientMu.RLock()
	_, alreadyInGame := h.clientToGame[client]
	h.clientMu.RUnlock()
	if alreadyInGame {
		log.Printf("Client %s tried to join game but is already associated with one.", client.ID)
		h.sendJoinError(client, "Already in a game or lobby.") // Use specific join error
		return
	}

	var payload protocol.JoinGamePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		log.Printf("Error unmarshalling join_game payload from client %s: %v", client.ID, err)
		h.sendJoinError(client, "Invalid join_game message format.")
		return
	}
	if payload.Name == "" {
		log.Printf("Client %s tried to join with an empty name.", client.ID)
		h.sendJoinError(client, "Name cannot be empty.")
		return
	}
	if payload.GameCode == "" {
		log.Printf("Client %s tried to join without a game code.", client.ID)
		h.sendJoinError(client, "Game code cannot be empty.")
		return
	}
	if payload.DesiredTeam != 1 && payload.DesiredTeam != 2 {
		log.Printf("Client %s tried to join with an invalid desired team: %d", client.ID, payload.DesiredTeam)
		h.sendJoinError(client, "Invalid desired team.")
		return
	}
	gameCode := strings.ToUpper(payload.GameCode) // Normalize game code

	h.lobbyMu.Lock() // Lock lobbies for modification
	lobby, lobbyExists := h.lobbies[gameCode]
	if !lobbyExists {
		h.lobbyMu.Unlock()
		log.Printf("Client %s tried to join non-existent lobby %s", client.ID, gameCode)
		h.sendJoinError(client, "Game code not found.")
		return
	}

	if len(lobby) >= 4 {
		h.lobbyMu.Unlock()
		log.Printf("Client %s tried to join full lobby %s", client.ID, gameCode)
		h.sendJoinError(client, "Game lobby is full.")
		return
	}

	// Check for duplicate name within this lobby
	for _, existingClient := range lobby {
		if existingClient.Name == payload.Name {
			h.lobbyMu.Unlock()
			log.Printf("Client %s tried to join lobby %s with duplicate name '%s'", client.ID, gameCode, payload.Name)
			h.sendJoinError(client, "Name already taken in this lobby.")
			return
		}
	}

	// Add client to lobby
	client.Name = payload.Name // Set name before adding to lobby list
	client.DesiredTeam = payload.DesiredTeam // Set desired team
	newLobby := append(lobby, client)
	h.lobbies[gameCode] = newLobby
	h.lobbyMu.Unlock() // Unlock lobbyMu after modification

	// Update client mapping
	h.clientMu.Lock()
	h.clientToGame[client] = gameCode
	h.clientMu.Unlock()

	log.Printf("Client %s (%s) joined lobby %s. Lobby size: %d", client.ID, client.Name, gameCode, len(newLobby))

	// Broadcast updated lobby state
	h.broadcastLobbyUpdate(gameCode, newLobby)

	// Check if lobby is full and start game
	if len(newLobby) == 4 {
		log.Printf("Lobby %s is full. Starting game...", gameCode)

		// Lock gameMu before modifying games map
		h.gameMu.Lock()
		// Lock lobbyMu to safely delete the lobby
		h.lobbyMu.Lock()

		// Double-check lobby exists and has 4 players before proceeding
		finalLobby, finalLobbyExists := h.lobbies[gameCode]
		if !finalLobbyExists || len(finalLobby) != 4 {
			// Should not happen if locks are correct, but good safeguard
			log.Printf("Error: Lobby %s state changed unexpectedly before game start. Aborting start.", gameCode)
			h.lobbyMu.Unlock()
			h.gameMu.Unlock()
			errorMsgBytes, _ := protocol.NewMessage("error", protocol.ErrorPayload{Message: "Failed to start game due to internal error."})
			h.broadcastToLobby(gameCode, errorMsgBytes)
			return
		}

		// Create and start the game
		targetScore := 31 // Is this needed?
		gamePlayers := convertClientsToGamePlayers(finalLobby) // Use finalLobby slice
		newGame := game.NewGame(gamePlayers, targetScore) 
		h.games[gameCode] = newGame // Add to games map using gameCode

		// Remove the lobby now that the game is created
		delete(h.lobbies, gameCode)

		h.lobbyMu.Unlock() // Unlock lobbyMu
		h.gameMu.Unlock() // Unlock gameMu

		log.Printf("Game instance created for code %s with ID %s. Players: %v", gameCode, newGame.ID, playerNames(finalLobby))

		// Start the game loop/first round in a goroutine
		// This function should handle sending game_start, deal_hand, etc.
		go newGame.StartGameLoop(h.sendMessageToClient) // Pass the callback
	}
}

// handleGameAction forwards actions like play_card or declare to the correct game instance.
func (h *Hub) handleGameAction(client *Client, msg protocol.Message) {
	h.clientMu.RLock()
	gameCode, inGame := h.clientToGame[client]
	h.clientMu.RUnlock()

	if !inGame {
		log.Printf("Received '%s' from client %s not in any game/lobby.", msg.Type, client.ID)
		h.sendErrorToClient(client, "You are not in an active game or lobby.")
		return
	}

	h.gameMu.RLock()
	gameInstance, gameExists := h.games[gameCode]
	h.gameMu.RUnlock()

	if !gameExists {
		// Could happen if message arrives after game ended/player disconnected but before unregister processed fully
		// Or if they were only in a lobby
		log.Printf("Received '%s' from client %s for game code %s, but game instance not found (maybe still in lobby or game ended?).", msg.Type, client.ID, gameCode)
		h.sendErrorToClient(client, "Game not found or not active.")
		return
	}

	log.Printf("Forwarding '%s' from client %s to game %s (Instance ID: %s)", msg.Type, client.ID, gameCode, gameInstance.ID)
	// Forward the message to the game instance for processing
	gameInstance.HandlePlayerAction(client.ID, msg) // Pass client ID and message
}


// Helper to get player names for logging
func playerNames(players []*Client) []string {
	names := make([]string, len(players))
	for i, p := range players {
		if p != nil { // Add nil check
			names[i] = p.Name
		} else {
			names[i] = "<nil>"
		}
	}
	return names
}

// Helper to convert server Clients to game Players
func convertClientsToGamePlayers(clients []*Client) [4]*shared.Player {
	if len(clients) != 4 {
		log.Printf("Error: convertClientsToGamePlayers called with %d clients, expected 4", len(clients))
		return [4]*shared.Player{} // Return empty array or handle error appropriately
	}
	var gamePlayers [4]*shared.Player
	for i, c := range clients {
		if c == nil {
			log.Printf("Error: Nil client found at index %d during conversion", i)
			// Handle error: return empty, panic, or skip? Returning empty for now.
			return [4]*shared.Player{}
		}
		gamePlayers[i] = shared.NewPlayer(c.ID, c.Name, c.DesiredTeam)
	}
	return gamePlayers
}

// sendMessageToClient allows the game logic to send messages back via the hub/client.
// This is passed as a callback to the game instance.
func (h *Hub) sendMessageToClient(clientID string, message []byte) {
	h.clientMu.RLock()
	// Find the client pointer using the ID
	var targetClient *Client
	for client := range h.clients {
		if client.ID == clientID {
			targetClient = client
			break
		}
	}
	h.clientMu.RUnlock() // Unlock after finding the client

	if targetClient != nil {
		// Use a non-blocking send with select to avoid blocking the hub/game goroutine
		select {
		case targetClient.send <- message:
			// Message sent successfully
		default:
			// Channel is blocked or closed, assume client disconnected
			log.Printf("Failed to send message to client %s (channel full or closed), initiating cleanup.", clientID)
			// Trigger client cleanup by sending to unregister channel
			// Use a goroutine to avoid potential deadlock if Run loop is busy
			go func() {
				// Check if client is still considered connected before unregistering
				h.clientMu.RLock()
				_, stillConnected := h.clients[targetClient]
				h.clientMu.RUnlock()
				if stillConnected {
					h.unregister <- targetClient
				}
			}()
		}
	} else {
		log.Printf("Could not find client %s to send message (already disconnected?).", clientID)
	}
}


// broadcastToLobby sends a message to all clients currently in a specific lobby.
func (h *Hub) broadcastToLobby(gameCode string, message []byte) {
	h.lobbyMu.RLock()
	lobby, exists := h.lobbies[gameCode]
	if !exists {
		h.lobbyMu.RUnlock()
		log.Printf("Warning: Tried to broadcast to non-existent lobby %s", gameCode)
		return
	}
	// Create a copy of the slice to avoid holding lock during send
	clientsToSend := make([]*Client, len(lobby))
	copy(clientsToSend, lobby)
	h.lobbyMu.RUnlock()

	log.Printf("Broadcasting message to %d clients in lobby %s", len(clientsToSend), gameCode)
	for _, client := range clientsToSend {
		if client != nil {
			select {
			case client.send <- message:
			default:
				log.Printf("Failed to send lobby message to client %s (channel full or closed)", client.ID)
				// Consider triggering unregister for this client
				go func(c *Client) {
					h.clientMu.RLock()
					_, stillConnected := h.clients[c]
					h.clientMu.RUnlock()
					if stillConnected {
						h.unregister <- c
					}
				}(client)
			}
		}
	}
}

// broadcastLobbyUpdate sends the current list of players in the lobby.
func (h *Hub) broadcastLobbyUpdate(gameCode string, lobby []*Client) {
	playerInfos := make([]protocol.PlayerInfo, len(lobby))
	for i, c := range lobby {
		if c != nil {
			playerInfos[i] = protocol.PlayerInfo{ID: c.ID, Name: c.Name}
		}
	}
	payload := protocol.LobbyUpdatePayload{Players: playerInfos}
	msgBytes, err := protocol.NewMessage("lobby_update", payload)
	if err != nil {
		log.Printf("Error creating lobby_update message for lobby %s: %v", gameCode, err)
		return
	}
	h.broadcastToLobby(gameCode, msgBytes)
}


// sendErrorToClient sends a generic error message to a specific client.
func (h *Hub) sendErrorToClient(client *Client, errorMsg string) {
	payload := protocol.ErrorPayload{Message: errorMsg}
	msgBytes, err := protocol.NewMessage("error", payload)
	if err != nil {
		log.Printf("Error creating error message for client %s: %v", client.ID, err)
		return
	}
	// Use sendMessageToClient which handles finding the client and non-blocking send
	h.sendMessageToClient(client.ID, msgBytes)
}

// sendJoinError sends a specific join error message to a client.
func (h *Hub) sendJoinError(client *Client, errorMsg string) {
	payload := protocol.JoinErrorPayload{Message: errorMsg} 
	msgBytes, err := protocol.NewMessage("join_error", payload)
	if err != nil {
		log.Printf("Error creating join_error message for client %s: %v", client.ID, err)
		return
	}
	h.sendMessageToClient(client.ID, msgBytes)
}