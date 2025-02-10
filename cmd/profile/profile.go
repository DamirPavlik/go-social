package profile

import (
	"chat-go-htmx/cmd/render"
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

type ProfileData struct {
	ID             int
	Username       string
	Friendship     string
	ProfilePicture string
}

func GetProfile(c echo.Context, db *sql.DB, tmpl *template.Template) error {
	userID, _ := GetCurrentUser(c, db)
	profileId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return render.RenderTemplate(c, tmpl, "error", "invalid profile id")
	}

	var profile ProfileData
	err = db.QueryRow("SELECT id, username, profile_picture FROM users WHERE id = $1", profileId).Scan(&profile.ID, &profile.Username, &profile.ProfilePicture)
	if err != nil {
		return render.RenderTemplate(c, tmpl, "error", "user not found")
	}

	var status sql.NullString
	err = db.QueryRow("SELECT status FROM friend_request WHERE (sender_id = $1 AND reciever_id = $2) OR (sender_id = $2 AND reciever_id = $1)", userID, profileId).Scan(&status)

	if err == sql.ErrNoRows {
		profile.Friendship = "none"
	} else if err == nil {
		if status.String == "accepted" {
			profile.Friendship = "friends"
		} else if status.String == "pending" {
			profile.Friendship = "pending"
		} else if status.String == "declined" {
			profile.Friendship = "declined"
		}
	} else {
		return render.RenderTemplate(c, tmpl, "error", "db err")
	}

	return render.RenderTemplate(c, tmpl, "profile", profile)
}

func GetCurrentUser(c echo.Context, db *sql.DB) (userId int, username string) {
	cookie, err := c.Cookie("session")
	if err != nil {
		log.Println("err getting session: ", err)
		return
	}

	username = cookie.Value
	var id int
	err = db.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&id)
	if err != nil {
		log.Println("err getting user", err)
		return
	}

	return id, username
}

func GetUsernameById(c echo.Context, db *sql.DB) error {
	userId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	var username string
	err = db.QueryRow("SELECT username FROM users WHERE id = $1", userId).Scan(&username)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
	}

	return c.JSON(http.StatusOK, map[string]string{"success": username})
}

func GetCurrentUserIdJSON(c echo.Context, db *sql.DB) error {
	cookie, err := c.Cookie("session")
	if err != nil {
		log.Println("err getting session: ", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "err getting session"})
	}
	username := cookie.Value
	var id int
	err = db.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&id)
	if err != nil {
		log.Println("err getting user", err)
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
		}
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "error getting user"})
	}

	return c.JSON(http.StatusOK, map[string]int{"success": id})
}

func GetMyProfile(c echo.Context, db *sql.DB, tmpl *template.Template) error {
	userId, _ := GetCurrentUser(c, db)

	type User struct {
		Username       string
		Email          string
		ProfilePicture string
		CreatedAt      string
	}

	var rawCreatedAt time.Time
	var user User
	err := db.QueryRow("SELECT username, email, profile_picture, created_at FROM users WHERE id = $1", userId).
		Scan(&user.Username, &user.Email, &user.ProfilePicture, &rawCreatedAt)

	if err != nil {
		log.Println("Error getting user:", err)
		if err == sql.ErrNoRows {
			return render.RenderTemplate(c, tmpl, "error", "User not found")
		}
		return render.RenderTemplate(c, tmpl, "error", "Error getting user")
	}

	user.CreatedAt = rawCreatedAt.Format("02 Jan 2006, 15:04:05 MST")

	return render.RenderTemplate(c, tmpl, "my_profile", user)
}
