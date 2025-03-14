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

func SendFriendRequestAfterDelcine(c echo.Context, db *sql.DB, tmpl *template.Template) error {
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
	_, err = db.Exec("DELETE FROM friend_request WHERE (sender_id = $1 AND reciever_id = $2) OR (sender_id = $2 AND reciever_id = $1)", senderId, recieverId)
	if err != nil {
		log.Println("Error deleting friend request:", err)
		return render.RenderTemplate(c, tmpl, "error", "error deleting the old request")
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

func DeclineFriendRequest(c echo.Context, db *sql.DB, tmpl *template.Template) error {
	recieverId, _ := GetCurrentUser(c, db)
	if recieverId == 0 {
		return render.RenderTemplate(c, tmpl, "error", "invalid sender id")
	}

	senderId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return render.RenderTemplate(c, tmpl, "error", "invalid reciever id")
	}

	_, err = db.Exec(`UPDATE friend_request SET status = 'declined' WHERE sender_id = $1 AND reciever_id = $2`, senderId, recieverId)
	if err != nil {
		return render.RenderTemplate(c, tmpl, "error", "err updating the friend req status")
	}

	return render.RenderTemplate(c, tmpl, "success", "friend request declined")
}

func GetAllFriends(c echo.Context, db *sql.DB, tmpl *template.Template) error {
	currentUserId, _ := GetCurrentUser(c, db)
	friends := []struct {
		ID             int
		Username       string
		ProfilePicture string
	}{}

	rows, err := db.Query(`
		SELECT u.id, u.username, u.profile_picture
		FROM friends f
		JOIN users u ON
			(f.user1 = $1 AND f.user2 = u.id) OR
			(f.user2 = $1 AND f.user1 = u.id)
	`, currentUserId)
	if err != nil {
		err = render.RenderTemplate(c, tmpl, "error", "db err")
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("Template error: %v", err))
		}
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var friend struct {
			ID             int
			Username       string
			ProfilePicture string
		}
		if err := rows.Scan(&friend.ID, &friend.Username, &friend.ProfilePicture); err != nil {
			return render.RenderTemplate(c, tmpl, "error", "err scanning")
		}
		friends = append(friends, friend)
	}

	if err := rows.Err(); err != nil {
		return render.RenderTemplate(c, tmpl, "error", "err rows")
	}

	err = render.RenderTemplate(c, tmpl, "friends", friends)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Template error: %v", err))
	}
	return nil
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

func RemoveFriend(c echo.Context, db *sql.DB, tmpl *template.Template) error {
	currentUserId, _ := GetCurrentUser(c, db)
	userId, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return render.RenderTemplate(c, tmpl, "error", "invalid user id")
	}

	_, err = db.Exec("DELETE FROM friend_request WHERE (sender_id = $1 AND reciever_id = $2) OR (sender_id = $2 AND reciever_id = $1)", currentUserId, userId)
	if err != nil {
		log.Println("err friend requests: ", err)
		return render.RenderTemplate(c, tmpl, "error", "err deleting friend requests")
	}

	_, err = db.Exec(`DELETE FROM friends WHERE (user1 = $1 AND user2 = $2) OR (user1 = $2 AND user2 = $1)`, currentUserId, userId)
	if err != nil {
		log.Println("err deleting user: ", err)
		return render.RenderTemplate(c, tmpl, "error", "err deleting user")
	}

	return render.RenderTemplate(c, tmpl, "reload", "")
}
