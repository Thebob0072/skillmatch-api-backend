package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

var (
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// In production, check origin properly
		return true
	},
}

// WebSocketManager manages all WebSocket connections
type WebSocketManager struct {
	clients    map[int]*Client // userID -> Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan BroadcastMessage
	mu         sync.RWMutex
}

// Client represents a WebSocket client connection
type Client struct {
	userID        int
	conn          *websocket.Conn
	send          chan []byte
	manager       *WebSocketManager
	lastPing      time.Time
	authenticated bool
	authTimeout   time.Time
}

// BroadcastMessage represents a message to broadcast to specific users
type BroadcastMessage struct {
	userIDs []int
	message []byte
}

// Global WebSocket manager instance
var wsManager *WebSocketManager

// InitWebSocketManager initializes the global WebSocket manager
func InitWebSocketManager() {
	wsManager = &WebSocketManager{
		clients:    make(map[int]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan BroadcastMessage, 256),
	}
	go wsManager.Run()
}

// Run starts the WebSocket manager event loop
func (m *WebSocketManager) Run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-m.register:
			m.mu.Lock()
			// Only register authenticated clients
			if client.authenticated && client.userID > 0 {
				// If user already has a connection, close the old one
				if existingClient, ok := m.clients[client.userID]; ok {
					close(existingClient.send)
					existingClient.conn.Close()
				}
				m.clients[client.userID] = client
				log.Printf("WebSocket: User %d connected (total: %d)", client.userID, len(m.clients))
			}
			m.mu.Unlock()

		case client := <-m.unregister:
			m.mu.Lock()
			if _, ok := m.clients[client.userID]; ok {
				delete(m.clients, client.userID)
				close(client.send)
			}
			m.mu.Unlock()
			log.Printf("WebSocket: User %d disconnected (total: %d)", client.userID, len(m.clients))

		case broadcast := <-m.broadcast:
			m.mu.RLock()
			for _, userID := range broadcast.userIDs {
				if client, ok := m.clients[userID]; ok {
					select {
					case client.send <- broadcast.message:
					default:
						// Client's send buffer is full, disconnect them
						close(client.send)
						delete(m.clients, client.userID)
					}
				}
			}
			m.mu.RUnlock()

		case <-ticker.C:
			// Ping all clients to keep connections alive
			m.mu.RLock()
			for _, client := range m.clients {
				if time.Since(client.lastPing) > 60*time.Second {
					// Client hasn't responded in 60 seconds, disconnect
					go func(c *Client) {
						m.unregister <- c
					}(client)
				}
			}
			m.mu.RUnlock()
		}
	}
}

// BroadcastToUser sends a message to a specific user
func (m *WebSocketManager) BroadcastToUser(userID int, message WebSocketMessage) {
	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("WebSocket: Failed to marshal message: %v", err)
		return
	}

	m.broadcast <- BroadcastMessage{
		userIDs: []int{userID},
		message: jsonData,
	}
}

// BroadcastToUsers sends a message to multiple users
func (m *WebSocketManager) BroadcastToUsers(userIDs []int, message WebSocketMessage) {
	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("WebSocket: Failed to marshal message: %v", err)
		return
	}

	m.broadcast <- BroadcastMessage{
		userIDs: userIDs,
		message: jsonData,
	}
}

// IsUserOnline checks if a user is currently connected
func (m *WebSocketManager) IsUserOnline(userID int) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.clients[userID]
	return ok
}

// GetOnlineUserCount returns the number of connected users
func (m *WebSocketManager) GetOnlineUserCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clients)
}

// HandleWebSocket handles WebSocket connection upgrades
// GET /ws (no token required initially - authenticate via message)
func HandleWebSocket(c *gin.Context) {
	// Upgrade HTTP connection to WebSocket immediately (no auth check yet)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Create unauthenticated client
	client := &Client{
		userID:        0, // Not authenticated yet
		conn:          conn,
		send:          make(chan []byte, 256),
		manager:       wsManager,
		lastPing:      time.Now(),
		authenticated: false,
		authTimeout:   time.Now().Add(10 * time.Second), // Must auth within 10 seconds
	}

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// readPump reads messages from the WebSocket connection
func (c *Client) readPump() {
	defer func() {
		c.manager.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.lastPing = time.Now()
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming messages (e.g., typing indicators)
		var wsMsg WebSocketMessage
		if err := json.Unmarshal(message, &wsMsg); err == nil {
			c.handleIncomingMessage(wsMsg)
		}
	}
}

// writePump writes messages to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Channel closed
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleIncomingMessage processes messages received from the client
func (c *Client) handleIncomingMessage(msg WebSocketMessage) {
	switch msg.Type {
	case "auth":
		// Handle authentication
		if !c.authenticated {
			if data, ok := msg.Payload.(map[string]interface{}); ok {
				if tokenString, ok := data["token"].(string); ok {
					// Validate JWT token
					token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
						if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
							return nil, fmt.Errorf("unexpected signing method")
						}
						return jwtSecret, nil
					})

					if err == nil && token.Valid {
						if claims, ok := token.Claims.(jwt.MapClaims); ok {
							var userID int
							// Try "sub" first (standard JWT format - string)
							if subStr, ok := claims["sub"].(string); ok {
								userIDInt, err := strconv.Atoi(subStr)
								if err == nil {
									userID = userIDInt
								} else {
									log.Printf("❌ Cannot parse user_id from sub: %s", subStr)
								}
							} else if userIDFloat, ok := claims["user_id"].(float64); ok {
								// Fallback to old format
								userID = int(userIDFloat)
							}

							if userID > 0 {
								c.userID = userID
								c.authenticated = true

								// Register this client
								wsManager.register <- c

								// Send success response
								response := WebSocketMessage{
									Type: "auth_success",
									Payload: map[string]interface{}{
										"user_id": userID,
										"message": "Authentication successful",
									},
								}
								jsonData, _ := json.Marshal(response)
								c.send <- jsonData

								log.Printf("✅ WebSocket: User %d authenticated", userID)
								return
							}
						}
					}
				}
			}

			// Auth failed
			response := WebSocketMessage{
				Type:    "auth_error",
				Payload: map[string]interface{}{"error": "Invalid token"},
			}
			jsonData, _ := json.Marshal(response)
			c.send <- jsonData
			c.conn.Close()
		}

	case "typing":
		// Only handle if authenticated
		if !c.authenticated {
			return
		}
		// Handle typing indicator
		if data, ok := msg.Payload.(map[string]interface{}); ok {
			if convID, ok := data["conversation_id"].(float64); ok {
				// Get the other user in the conversation
				var otherUserID int
				db.QueryRow(`
					SELECT CASE 
						WHEN user1_id = $1 THEN user2_id 
						ELSE user1_id 
					END 
					FROM conversations WHERE id = $2
				`, c.userID, int(convID)).Scan(&otherUserID)

				if otherUserID > 0 {
					wsManager.BroadcastToUser(otherUserID, WebSocketMessage{
						Type: "typing",
						Payload: TypingIndicator{
							ConversationID: int(convID),
							UserID:         c.userID,
							IsTyping:       data["is_typing"].(bool),
						},
					})
				}
			}
		}

	case "ping":
		// Respond with pong
		c.lastPing = time.Now()
		response := WebSocketMessage{
			Type:    "pong",
			Payload: map[string]interface{}{"timestamp": time.Now().Unix()},
		}
		jsonData, _ := json.Marshal(response)
		c.send <- jsonData
	}
}
