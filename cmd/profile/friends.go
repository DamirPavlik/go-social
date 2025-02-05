package profile

import (
	"chat-go-htmx/cmd/render"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func SendFriendRequest(c echo.Context, db *sql.DB, tmpl *template.Template) error {
	senderId, _ := GetCurrentUser(c, db)
	if senderId == 0 {
		return render.RenderTemplate(c, tmpl, "error", "invalid sender id")
	}

	recieverId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return render.RenderTemplate(c, tmpl, "error", "invalid reciever id")
	}

	if senderId == recieverId {
		log.Println("retard: ", err)
		return render.RenderTemplate(c, tmpl, "error", "are you dumb you can't send req to yourself")
	}

	_, err = db.Exec("INSERT INTO friend_request (sender_id, reciever_id, status) VALUES ($1, $2, 'pending') ON CONFLICT DO NOTHING", senderId, recieverId)
	if err != nil {
		log.Println("err sending friend request: ", err)
		return render.RenderTemplate(c, tmpl, "error", "db err")
	}

	return render.RenderTemplate(c, tmpl, "success", "friend reqest sent")
}

func AcceptFriendRequest(c echo.Context, db *sql.DB, tmpl *template.Template) error {
	recieverId, _ := GetCurrentUser(c, db)
	if recieverId == 0 {
		return render.RenderTemplate(c, tmpl, "error", "invalid sender id")
	}

	senderId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return render.RenderTemplate(c, tmpl, "error", "invalid reciever id")
	}

	_, err = db.Exec(`UPDATE friend_request SET status = 'accepted' WHERE sender_id = $1 AND reciever_id = $2`, senderId, recieverId)
	if err != nil {
		return render.RenderTemplate(c, tmpl, "error", "err updating the friend req status")
	}

	_, err = db.Exec("INSERT INTO friends (user1, user2) VALUES ($1, $2)", senderId, recieverId)

	if err != nil {
		return render.RenderTemplate(c, tmpl, "error", "db err")
	}
	return render.RenderTemplate(c, tmpl, "success", "friend request accepted")
}

func GetAllFriendRequests(c echo.Context, db *sql.DB, tmpl *template.Template) error {
	currentUserId, _ := GetCurrentUser(c, db)

	friendRequests := []struct {
		ID       int
		Username string
	}{}

	rows, err := db.Query(`
		SELECT users.id, users.username 
		FROM friend_request 
		JOIN users ON friend_request.sender_id = users.id 
		WHERE friend_request.reciever_id = $1 AND friend_request.status = 'pending'`,
		currentUserId)

	if err != nil {
		log.Println("error db: ", err)
		err = render.RenderTemplate(c, tmpl, "error", "db err")
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("Template error: %v", err))
		}
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var user struct {
			ID       int
			Username string
		}
		if err := rows.Scan(&user.ID, &user.Username); err != nil {
			return render.RenderTemplate(c, tmpl, "error", "err scanning")
		}
		friendRequests = append(friendRequests, user)
	}

	if err := rows.Err(); err != nil {
		return render.RenderTemplate(c, tmpl, "error", "err rows")
	}

	err = render.RenderTemplate(c, tmpl, "friend_requests", friendRequests)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Template error: %v", err))
	}
	return nil
}
