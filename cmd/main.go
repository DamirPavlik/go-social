package main

import (
	"chat-go-htmx/cmd/auth"
	"chat-go-htmx/cmd/chat"
	"chat-go-htmx/cmd/middleware"
	"chat-go-htmx/cmd/posts"
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

var (
	db        *sql.DB
	viewsPath = "../views/"
	templates map[string]*template.Template
)

func loadEnv() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("Warning: .env file not found")
	}
}

func initDB() {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is missing in the environment variables")
	}
	var err error
	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}
}

func loadTemplates() {
	templates = map[string]*template.Template{
		"auth":           template.Must(template.ParseFiles(viewsPath + "templates/auth.html")),
		"search":         template.Must(template.ParseFiles(viewsPath + "templates/search_results.html")),
		"profile":        template.Must(template.ParseFiles(viewsPath + "templates/profile.html")),
		"feed":           template.Must(template.ParseFiles(viewsPath + "templates/posts_feed.html")),
		"myProfile":      template.Must(template.ParseFiles(viewsPath + "my_profile.html")),
		"friendRequests": template.Must(template.ParseFiles(viewsPath + "templates/friend_requests.html")),
		"friends":        template.Must(template.ParseFiles(viewsPath + "templates/friends.html")),
		"posts":          template.Must(template.ParseFiles(viewsPath + "templates/posts.html")),
	}
}

func serveAssets(e *echo.Echo) {
	e.Static("/assets", "../assets")
	e.Static("/profile_pictures", "../uploads/profile_pictures")
	e.Static("/uploads/posts", "../uploads/posts")
}

func setupRoutes(e *echo.Echo, chatManager *chat.ChatManager) {
	// Home page
	e.GET("/", func(c echo.Context) error {
		cookie, err := c.Cookie("session")
		if err != nil || cookie.Value == "" {
			return c.Redirect(http.StatusSeeOther, "/register")
		}
		return c.File(viewsPath + "index.html")
	})

	// Authentication routes
	e.GET("/register", func(c echo.Context) error { return c.File(viewsPath + "register.html") })
	e.GET("/login", func(c echo.Context) error {
		if cookie, err := c.Cookie("session"); err == nil && cookie.Value != "" {
			return c.Redirect(http.StatusSeeOther, "/")
		}
		return c.File(viewsPath + "login.html")
	})
	e.POST("/register", func(c echo.Context) error { return auth.RegisterUser(c, db, templates["auth"]) })
	e.POST("/login", func(c echo.Context) error { return auth.LoginUser(c, db, templates["auth"]) })
	e.POST("/logout", func(c echo.Context) error { return auth.LogoutUser(c, templates["auth"]) })

	authGroup := e.Group("")
	authGroup.Use(middleware.AuthMiddleware())

	// Search
	authGroup.GET("/search", func(c echo.Context) error { return search.SearchUsers(c, db, templates["search"]) })

	// Profile
	authGroup.GET("/user-username/:id", func(c echo.Context) error { return profile.GetUsernameById(c, db) })
	authGroup.GET("/current-user-id", func(c echo.Context) error { return profile.GetCurrentUserIdJSON(c, db) })
	authGroup.GET("/profile/:id", func(c echo.Context) error { return profile.GetProfile(c, db, templates["profile"]) })
	authGroup.GET("/my-profile", func(c echo.Context) error { return profile.GetMyProfile(c, db, templates["myProfile"]) })
	authGroup.POST("/edit-profile", func(c echo.Context) error { return profile.EditMyProfile(c, db, templates["myProfile"]) })

	// Friend Requests
	authGroup.GET("/friend-requests", func(c echo.Context) error { return profile.GetAllFriendRequests(c, db, templates["friendRequests"]) })
	authGroup.GET("/friends", func(c echo.Context) error { return profile.GetAllFriends(c, db, templates["friends"]) })
	authGroup.POST("/profile/:id/add", func(c echo.Context) error { return profile.SendFriendRequest(c, db, templates["profile"]) })
	authGroup.POST("/profile/:id/add-after-decline", func(c echo.Context) error { return profile.SendFriendRequestAfterDelcine(c, db, templates["profile"]) })
	authGroup.POST("/accept/:id", func(c echo.Context) error { return profile.AcceptFriendRequest(c, db, templates["profile"]) })
	authGroup.POST("/decline/:id", func(c echo.Context) error { return profile.DeclineFriendRequest(c, db, templates["profile"]) })
	authGroup.POST("/remove-friend/:id", func(c echo.Context) error { return profile.RemoveFriend(c, db, templates["friends"]) })

	// Posts
	authGroup.POST("/post", func(c echo.Context) error { return posts.CreatePost(c, db, templates["posts"]) })
	authGroup.GET("/profile/:id/posts", func(c echo.Context) error { return posts.GetUserPosts(c, db, templates["posts"]) })
	authGroup.GET("/friends-feed", func(c echo.Context) error { return posts.GetFriendsPosts(c, db, templates["feed"]) })
	authGroup.GET("/current-user-posts", func(c echo.Context) error { return posts.GetCurrentUsersPosts(c, db, templates["myProfile"]) })
	authGroup.POST("/post/:id/like", func(c echo.Context) error { return posts.LikePost(c, db, templates["posts"]) })
	authGroup.POST("/post/:id/unlike", func(c echo.Context) error { return posts.UnlikePost(c, db, templates["posts"]) })
	authGroup.POST("/post/:id/comment", func(c echo.Context) error { return posts.CommentOnPost(c, db, templates["posts"]) })

	// Feed Likes
	authGroup.POST("/feed-post/:id/like", func(c echo.Context) error { return posts.LikePost(c, db, templates["feed"]) })
	authGroup.POST("/feed-post/:id/unlike", func(c echo.Context) error { return posts.UnlikePost(c, db, templates["feed"]) })

	// Chat
	authGroup.GET("/chat/:id", func(c echo.Context) error { return chatManager.HandleChat(c) })
	go chatManager.HandleMessage()
}

func main() {
	loadEnv()
	initDB()
	loadTemplates()

	e := echo.New()
	chatManager := chat.NewChatManager(db)

	setupRoutes(e, chatManager)
	serveAssets(e)

	e.Logger.Fatal(e.Start(":8080"))
}
