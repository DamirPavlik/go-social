package main

import (
	"chat-go-htmx/cmd/auth"
	"chat-go-htmx/cmd/chat"
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

var db *sql.DB
var viewsPath = "../views/"

func initDB(dbUrl string) {
	var err error
	db, err = sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal(err)
	}
}

func serveAssets(e *echo.Echo) {
	e.Static("/assets", "../assets")
	e.Static("/profile_pictures", "../uploads/profile_pictures")
	e.Static("/uploads/posts", "../uploads/posts")
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
	tmplFeed := template.Must(template.ParseFiles(viewsPath + "templates/posts_feed.html"))
	tmplMyProfile := template.Must(template.ParseFiles(viewsPath + "my_profile.html"))
	tmplFriendRequests := template.Must(template.ParseFiles(viewsPath + "templates/friend_requests.html"))
	tmplFriends := template.Must(template.ParseFiles(viewsPath + "templates/friends.html"))
	tmplPosts := template.Must(template.ParseFiles(viewsPath + "templates/posts.html"))

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

	e.GET("/user-username/:id", func(c echo.Context) error {
		return profile.GetUsernameById(c, db)
	})

	e.GET("/current-user-id", func(c echo.Context) error {
		return profile.GetCurrentUserIdJSON(c, db)
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

	e.POST("/post", func(c echo.Context) error {
		return posts.CreatePost(c, db, tmplPosts)
	})

	e.GET("/profile/:id/posts", func(c echo.Context) error {
		return posts.GetUserPosts(c, db, tmplPosts)
	})

	e.POST("/post/:id/like", func(c echo.Context) error {
		return posts.LikePost(c, db, tmplPosts)
	})

	e.POST("/feed-post/:id/like", func(c echo.Context) error {
		return posts.LikePost(c, db, tmplFeed)
	})

	e.POST("/feed-post/:id/unlike", func(c echo.Context) error {
		return posts.UnlikePost(c, db, tmplFeed)
	})

	e.POST("/post/:id/unlike", func(c echo.Context) error {
		return posts.UnlikePost(c, db, tmplPosts)
	})

	e.POST("/post/:id/comment", func(c echo.Context) error {
		return posts.CommentOnPost(c, db, tmplPosts)
	})

	e.GET("/my-profile", func(c echo.Context) error {
		return profile.GetMyProfile(c, db, tmplMyProfile)
	})

	e.GET("/friends-feed", func(c echo.Context) error {
		return posts.GetFriendsPosts(c, db, tmplFeed)
	})

	e.POST("/edit-profile", func(c echo.Context) error {
		return profile.EditMyProfile(c, db, tmplMyProfile)
	})

	e.GET("/current-user-posts", func(c echo.Context) error {
		return posts.GetCurrentUsersPosts(c, db, tmplMyProfile)
	})

	e.POST("/remove-friend/:id", func(c echo.Context) error {
		return profile.RemoveFriend(c, db, tmplFriends)
	})

	go chatManager.HandleMessage()
	serveAssets(e)

	e.Logger.Fatal(e.Start(":8080"))
}
