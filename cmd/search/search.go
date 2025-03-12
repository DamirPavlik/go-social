package search

import (
	"chat-go-htmx/cmd/profile"
	"database/sql"
	"html/template"
	"net/http"

	"github.com/labstack/echo/v4"
)

type User struct {
	ID       int
	Username string
}

func SearchUsers(c echo.Context, db *sql.DB, tmpl *template.Template) error {
	query := c.QueryParam("q")
	currentUserId, _ := profile.GetCurrentUser(c, db)

	if query == "" {
		return c.HTML(http.StatusOK, "")
	}

	rows, err := db.Query("SELECT id, username FROM users WHERE username ILIKE '%' || $1 || '%' AND id != $2 LIMIT 10", query, currentUserId)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Database error")
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username); err != nil {
			return c.String(http.StatusInternalServerError, "Error scanning rows")
		}
		users = append(users, u)
	}

	return tmpl.ExecuteTemplate(c.Response().Writer, "search_results", users)
}
