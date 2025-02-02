package auth

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func RegisterUser(c echo.Context, db *sql.DB) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	if username == "" || password == "" {
		return c.HTML(http.StatusBadRequest, `<div class="error">Username and password cannot be empty</div>`)
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 12)
	_, err := db.Exec("INSERT INTO users(username, password_hash, created_at) VALUES ($1, $2, $3)", username, hashedPassword, time.Now())

	if err != nil {
		return c.HTML(http.StatusInternalServerError, `<div class="error">Username already taken</div>`)
	}
	c.SetCookie(&http.Cookie{
		Name:  "session",
		Value: username,
		Path:  "/",
	})

	return c.Redirect(http.StatusSeeOther, "/")
}

func LoginUser(c echo.Context, db *sql.DB) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	if username == "" || password == "" {
		return c.HTML(http.StatusBadRequest, `<div class="error">Username and password cannot be empty</div>`)
	}

	var storedHash string
	err := db.QueryRow("SELECT password_hash FROM users WHERE username = $1", username).Scan(&storedHash)

	if err != nil {
		return c.HTML(http.StatusInternalServerError, `<div class="error">invalid username or password</div>`)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err != nil {
		return c.HTML(http.StatusInternalServerError, `<div class="error">invalid credentials</div>`)
	}

	c.SetCookie(&http.Cookie{
		Name:  "session",
		Value: username,
		Path:  "/",
	})

	return c.Redirect(http.StatusSeeOther, "/")
}
