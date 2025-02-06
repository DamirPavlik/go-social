package auth

import (
	"chat-go-htmx/cmd/render"
	"database/sql"
	"html/template"
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

func RegisterUser(c echo.Context, db *sql.DB, tmpl *template.Template) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	if username == "" || password == "" {
		return render.RenderTemplate(c, tmpl, "error", "Username and password cannot be empty")
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 12)
	_, err := db.Exec("INSERT INTO users(username, password_hash, created_at) VALUES ($1, $2, $3)", username, hashedPassword, time.Now())

	if err != nil {
		return render.RenderTemplate(c, tmpl, "error", "Username already taken")
	}

	c.SetCookie(&http.Cookie{
		Name:  "session",
		Value: username,
		Path:  "/",
	})

	return render.RenderTemplate(c, tmpl, "redirect", "/")
}

func LoginUser(c echo.Context, db *sql.DB, tmpl *template.Template) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	if username == "" || password == "" {
		return render.RenderTemplate(c, tmpl, "error", "Username and password cannot be empty")
	}

	var storedHash string
	err := db.QueryRow("SELECT password_hash FROM users WHERE username = $1", username).Scan(&storedHash)

	if err != nil {
		return render.RenderTemplate(c, tmpl, "error", "Invalid username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err != nil {
		return render.RenderTemplate(c, tmpl, "error", "Invalid credentials")
	}

	c.SetCookie(&http.Cookie{
		Name:  "session",
		Value: username,
		Path:  "/",
	})

	return render.RenderTemplate(c, tmpl, "redirect", "/")
}

func LogoutUser(c echo.Context, tmpl *template.Template) error {
	c.SetCookie(&http.Cookie{
		Name:  "session",
		Value: "",
		Path:  "/",
	})

	return render.RenderTemplate(c, tmpl, "redirect", "/register")
}
