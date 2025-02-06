package chat

import (
	"chat-go-htmx/cmd/profile"
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type Message struct {
	SenderID   int    `json:"sender_id"`
	RecieverID int    `json:"reciever_id"`
	Content    string `json:"content"`
}

type ChatManager struct {
	clients   map[*websocket.Conn]int
	broadcast chan Message
	upgrader  websocket.Upgrader
	mu        sync.Mutex
	db        *sql.DB
}

func NewChatManager(db *sql.DB) *ChatManager {
	return &ChatManager{
		clients:   make(map[*websocket.Conn]int),
		broadcast: make(chan Message),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		db: db,
	}
}

func (cm *ChatManager) HandleChat(c echo.Context) error {
	userId, _ := profile.GetCurrentUser(c, cm.db)
	recieverId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid user id")
	}

	conn, err := cm.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	cm.mu.Lock()
	cm.clients[conn] = userId
	cm.mu.Unlock()

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			cm.mu.Lock()
			delete(cm.clients, conn)
			cm.mu.Unlock()
			break
		}

		msg.SenderID = userId
		msg.RecieverID = recieverId

		_, err = cm.db.Exec("INSERT INTO messages (sender_id, reciever_id, content) VALUES ($1, $2, $3)", msg.SenderID, msg.RecieverID, msg.Content)
		if err != nil {
			log.Println("err saving message: ", err)
			continue
		}

		cm.broadcast <- msg
	}

	return nil
}

func (cm *ChatManager) HandleMessage() {
	for {
		msg := <-cm.broadcast
		cm.mu.Lock()
		for conn, userId := range cm.clients {
			if userId == msg.RecieverID || userId == msg.SenderID {
				conn.WriteJSON(msg)
			}
		}
		cm.mu.Unlock()
	}
}
