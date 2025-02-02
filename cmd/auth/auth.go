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

	c.Redirect(http.StatusOK, "/")
	return c.HTML(http.StatusOK, `<div class="success">User registered successfully!</div>`)
}

func LoginUser(c echo.Context, db *sql.DB) error {
	u := new(User)
	if err := c.Bind(u); err != nil {
		return err
	}

	var storedHash string
	err := db.QueryRow("SELECT password_hash FROM users WHERE username = $1", u.Username).Scan(*&storedHash)

	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid username or password"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(u.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	}

	c.SetCookie(&http.Cookie{
		Name:  "session",
		Value: u.Username,
		Path:  "/",
	})

	return c.JSON(http.StatusOK, map[string]string{"message": "user logged in succesfully"})
}
