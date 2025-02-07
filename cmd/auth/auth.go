package auth

import (
	"chat-go-htmx/cmd/render"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// type User struct {
// 	ID       int    `json:"id"`
// 	Username string `json:"username"`
// 	Password string `json:"password"`
// }

const uploadDir = "../uploads/profile_pictures/"

func RegisterUser(c echo.Context, db *sql.DB, tmpl *template.Template) error {
	username := c.FormValue("username")
	email := c.FormValue("email")
	password := c.FormValue("password")

	if username == "" || email == "" || password == "" {
		return render.RenderTemplate(c, tmpl, "error", "All fields are required")
	}

	file, err := c.FormFile("profile_picture")
	profilePicture := "default.jpg"

	if err == nil {
		src, err := file.Open()
		if err != nil {
			log.Println("err src: ", err)
			return render.RenderTemplate(c, tmpl, "error", "Error opening file")
		}
		defer src.Close()

		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			os.MkdirAll(uploadDir, 0755)
		}

		ext := filepath.Ext(file.Filename)
		filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		filePath := filepath.Join(uploadDir, filename)

		dst, err := os.Create(filePath)
		if err != nil {
			return render.RenderTemplate(c, tmpl, "error", "Error saving file")
		}
		defer dst.Close()

		if _, err = dst.ReadFrom(src); err != nil {
			return render.RenderTemplate(c, tmpl, "error", "Error writing file")
		}

		profilePicture = filename
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 12)

	_, err = db.Exec(
		"INSERT INTO users(username, email, password_hash, profile_picture, created_at) VALUES ($1, $2, $3, $4, $5)",
		username, email, hashedPassword, profilePicture, time.Now(),
	)

	if err != nil {
		return render.RenderTemplate(c, tmpl, "error", "Username or email already taken")
	}

	c.SetCookie(&http.Cookie{
		Name:  "session",
		Value: username,
		Path:  "/",
	})

	return render.RenderTemplate(c, tmpl, "redirect", "/")
}

func LoginUser(c echo.Context, db *sql.DB, tmpl *template.Template) error {
	identifier := c.FormValue("identifier")
	password := c.FormValue("password")

	if identifier == "" || password == "" {
		return render.RenderTemplate(c, tmpl, "error", "Username/email and password cannot be empty")
	}

	var storedHash string
	var username string
	err := db.QueryRow("SELECT password_hash, username FROM users WHERE username = $1 OR email = $1", identifier).Scan(&storedHash, &username)

	if err != nil {
		log.Println("err logging in: ", err)
		return render.RenderTemplate(c, tmpl, "error", "invalid username/email or password")
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
