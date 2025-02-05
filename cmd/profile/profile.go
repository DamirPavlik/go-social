package profile

import (
	"chat-go-htmx/cmd/render"
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type User struct {
	ID       int
	Username string
}

func GetProfile(c echo.Context, db *sql.DB, tmpl *template.Template) error {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid param id")
	}

	var user User
	err = db.QueryRow("SELECT id, username FROM users WHERE id = $1", userID).Scan(&user.ID, &user.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("jeb mi mater", err)
			return render.RenderTemplate(c, tmpl, "error", "user not found")
		}
		log.Println("koji kurac", err)
		return render.RenderTemplate(c, tmpl, "error", "db err")
	}

	return render.RenderTemplate(c, tmpl, "profile", user)
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
