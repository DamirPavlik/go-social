package posts

import (
	"chat-go-htmx/cmd/profile"
	"chat-go-htmx/cmd/render"
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"
)

type Comment struct {
	ID        int
	PostID    int
	UserID    int
	Content   string
	CreatedAt time.Time
}

type Post struct {
	ID        int
	UserID    int
	Content   string
	Image     string
	CreatedAt time.Time
	Comments  []Comment
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

		uploadDir := "../uploads/posts"
		os.MkdirAll(uploadDir, 0755)
		filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(file.Filename))
		imagePath = filepath.ToSlash(filepath.Join(uploadDir, filename))

		dst, err := os.Create(imagePath)
		if err != nil {
			return render.RenderTemplate(c, tmpl, "error", "error saving file")
		}
		defer dst.Close()

		io.Copy(dst, src)
	}

	_, err = db.Exec("INSERT INTO posts(user_id, content, image, created_at) VALUES ($1, $2, $3, $4)", userID, content, imagePath, time.Now())

	if err != nil {
		log.Println("error inserting", err)
		return render.RenderTemplate(c, tmpl, "error", "error creating post")
	}

	return render.RenderTemplate(c, tmpl, "success", "post created")
}

func GetUserPosts(c echo.Context, db *sql.DB, tmpl *template.Template) error {
	userId := c.Param("id")
	log.Println("user id: ", userId)

	rows, err := db.Query("SELECT id, content, image, created_at FROM posts WHERE user_id = $1 ORDER BY created_at DESC", userId)
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
			log.Println("err reading posts: ", err)
			return render.RenderTemplate(c, tmpl, "error", "err reading posts")
		}

		comments, err := GetCommentsForPost(db, post.ID)
		if err != nil {
			log.Println("error fetching comments: ", err)
			return render.RenderTemplate(c, tmpl, "error", "error fetching comments")
		}
		post.Comments = comments

		posts = append(posts, post)
	}

	err = render.RenderTemplate(c, tmpl, "user_posts", posts)
	if err != nil {
		log.Println("template error: ", err)
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Template error: %v", err))
	}
	return nil
}

func GetCommentsForPost(db *sql.DB, postID int) ([]Comment, error) {
	rows, err := db.Query("SELECT id, post_id, user_id, content, created_at FROM comments WHERE post_id = $1 ORDER BY created_at ASC", postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.CreatedAt)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, nil
}
