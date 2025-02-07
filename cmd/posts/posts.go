package posts

import (
	"chat-go-htmx/cmd/profile"
	"chat-go-htmx/cmd/render"
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"
)

type Post struct {
	ID        int
	UserID    int
	Content   string
	Image     string
	CreatedAt time.Time
}

func CreatePost(c echo.Context, db *sql.DB, tmpl *template.Template) error {
	userID, _ := profile.GetCurrentUser(c, db)
	content := c.FormValue("content")

	if content == "" {
		return render.RenderTemplate(c, tmpl, "error", "content can not be empty")
	}

	var imagePath string
	file, err := c.FormFile("image")
	if err == nil && file != nil {
		src, err := file.Open()
		if err != nil {
			return render.RenderTemplate(c, tmpl, "error", "error opening file")
		}
		defer src.Close()

		uploadDir := "uploads/posts"
		os.MkdirAll(uploadDir, 0755)
		filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(file.Filename))
		imagePath = filepath.Join(uploadDir, filename)

		dst, err := os.Create(imagePath)
		if err != nil {
			return render.RenderTemplate(c, tmpl, "error", "error saving file")
		}
		defer dst.Close()

		io.Copy(dst, src)
	}

	_, err = db.Exec("INSERT INTO posts(user_id, content, iamge, created_at) VALUES ($1, $2, $3, $4)", userID, content, imagePath, time.Now())

	if err != nil {
		log.Println("error inserting", err)
		return render.RenderTemplate(c, tmpl, "error", "error creating post")
	}

	return render.RenderTemplate(c, tmpl, "success", "post created")
}

func GetUserPosts(c echo.Context, db *sql.DB, tmpl *template.Template) error {
	userId := c.Param("id")

	rows, err := db.Query("SELECT id, content, image, created_at FROM posts WHERE user_Id = $1 ORDER BY created_at DESC", userId)
	if err != nil {
		log.Println("error fetching posts: ", err)
		return render.RenderTemplate(c, tmpl, "error", "failed to get posts")
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Content, &post.Image, &post.CreatedAt)
		if err != nil {
			return render.RenderTemplate(c, tmpl, "error", "err reading posts")
		}
		posts = append(posts, post)
	}

	return render.RenderTemplate(c, tmpl, "user_posts", posts)
}
