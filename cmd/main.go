package main

import (
	"chat-go-htmx/cmd/auth"
	"chat-go-htmx/cmd/chat"
	"chat-go-htmx/cmd/profile"
	"chat-go-htmx/cmd/search"
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

var db *sql.DB
var viewsPath = "../views/"

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

	chatManager := chat.NewChatManager(db)

	tmplAuth := template.Must(template.ParseFiles(viewsPath + "templates/auth.html"))
	tmplSearch := template.Must(template.ParseFiles(viewsPath + "templates/search_results.html"))
	tmplProfile := template.Must(template.ParseFiles(viewsPath + "templates/profile.html"))
	tmplFriendRequests := template.Must(template.ParseFiles(viewsPath + "templates/friend_requests.html"))
	tmplFriends := template.Must(template.ParseFiles(viewsPath + "templates/friends.html"))

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

	e.POST("/profile/:id/add-after-decline", func(c echo.Context) error {
		return profile.SendFriendRequestAfterDelcine(c, db, tmplProfile)
	})

	e.POST("/accept/:id", func(c echo.Context) error {
		return profile.AcceptFriendRequest(c, db, tmplProfile)
	})

	e.POST("/decline/:id", func(c echo.Context) error {
		return profile.DeclineFriendRequest(c, db, tmplProfile)
	})

	e.GET("/friend-requests", func(c echo.Context) error {
		return profile.GetAllFriendRequests(c, db, tmplFriendRequests)
	})

	e.GET("/friends", func(c echo.Context) error {
		return profile.GetAllFriends(c, db, tmplFriends)
	})

	e.GET("/chat/:id", func(c echo.Context) error {
		return chatManager.HandleChat(c)
	})

	go chatManager.HandleMessage()

	e.Static("/assets", "../assets")
	e.Logger.Fatal(e.Start(":8080"))
}
