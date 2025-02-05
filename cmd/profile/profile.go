package profile

import (
	"chat-go-htmx/cmd/render"
	"database/sql"
	"html/template"
	"log"
	"strconv"

	"github.com/labstack/echo/v4"
)

type ProfileData struct {
	ID         int
	Username   string
	Friendship string
}

func GetProfile(c echo.Context, db *sql.DB, tmpl *template.Template) error {
	userID, _ := GetCurrentUser(c, db)
	profileId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return render.RenderTemplate(c, tmpl, "error", "invalid profile id")
	}

	var profile ProfileData
	err = db.QueryRow("SELECT id, username FROM users WHERE id = $1", profileId).Scan(&profile.ID, &profile.Username)
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
	log.Println("status: ", status.String)

	return render.RenderTemplate(c, tmpl, "profile", profile)
	// userID, err := strconv.Atoi(c.Param("id"))

	// var user User
	// err = db.QueryRow("SELECT id, username FROM users WHERE id = $1", userID).Scan(&user.ID, &user.Username)
	// if err != nil {
	// 	if err == sql.ErrNoRows {
	// 		log.Println("jeb mi mater", err)
	// 		return render.RenderTemplate(c, tmpl, "error", "user not found")
	// 	}
	// 	log.Println("koji kurac", err)
	// 	return render.RenderTemplate(c, tmpl, "error", "db err")
	// }

	// return render.RenderTemplate(c, tmpl, "profile", user)
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
