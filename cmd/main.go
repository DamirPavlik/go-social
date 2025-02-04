package main

import (
	"chat-go-htmx/cmd/auth"
	"chat-go-htmx/cmd/profile"
	"chat-go-htmx/cmd/search"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

var db *sql.DB
var viewsPath = "../views/"

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan string)

func handleConnections(c echo.Context) error {
	fmt.Println("New WebSocket connection")

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		fmt.Println("err upgrading: ", err)
		return err
	}
	defer ws.Close()

	clients[ws] = true

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			delete(clients, ws)
			break
		}

		broadcast <- string(msg)
	}

	return nil
}

func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func initDB(dbUrl string) {
	var err error
	db, err = sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	godotenv.Load("../.env")
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("dburl is not found in the environment")
	}
	initDB(dbURL)

	e := echo.New()
	tmplAuth := template.Must(template.ParseFiles(viewsPath + "templates/auth.html"))
	tmplSearch := template.Must(template.ParseFiles(viewsPath + "templates/search_results.html"))
	tmplProfile := template.Must(template.ParseFiles(viewsPath + "templates/profile.html"))

	e.GET("/", func(c echo.Context) error {
		cookie, err := c.Cookie("session")
		if err != nil || cookie.Value == "" {
			return c.Redirect(http.StatusSeeOther, "/register")
		}
		return c.File(viewsPath + "index.html")
	})

	e.GET("/register", func(c echo.Context) error {
		return c.File(viewsPath + "register.html")
	})

	e.GET("/login", func(c echo.Context) error {
		cookie, err := c.Cookie("session")
		if err == nil && cookie.Value != "" {
			return c.Redirect(http.StatusSeeOther, "/")
		}
		return c.File(viewsPath + "login.html")
	})

	e.POST("/register", func(c echo.Context) error {
		cookie, err := c.Cookie("session")
		if err == nil && cookie.Value != "" {
			return c.Redirect(http.StatusSeeOther, "/")
		}
		return auth.RegisterUser(c, db, tmplAuth)
	})

	e.POST("/login", func(c echo.Context) error {
		return auth.LoginUser(c, db, tmplAuth)
	})

	e.POST("/logout", func(c echo.Context) error {
		return auth.LogoutUser(c, tmplAuth)
	})

	e.GET("/search", func(c echo.Context) error {
		return search.SearchUsers(c, db, tmplSearch)
	})

	e.GET("/profile/:id", func(c echo.Context) error {
		return profile.GetProfile(c, db, tmplProfile)
	})

	e.POST("/profile/:id/add", func(c echo.Context) error {
		return profile.SendFriendRequest(c, db, tmplProfile)
	})

	e.POST("/profile/:id/accept", func(c echo.Context) error {
		return profile.AcceptFriendRequest(c, db, tmplProfile)
	})

	e.GET("/ws", handleConnections)

	go handleMessages()

	e.Static("/assets", "../assets")
	e.Logger.Fatal(e.Start(":8080"))
}
