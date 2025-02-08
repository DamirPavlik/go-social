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
	Username  string
	Content   string
	CreatedAt time.Time
}

type Post struct {
	ID          int
	UserID      int
	Content     string
	Image       string
	CreatedAt   time.Time
	Comments    []Comment
	LikesCount  int
	LikedByUser bool
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
	userID, _ := profile.GetCurrentUser(c, db)
	profileUserID := c.Param("id")

	rows, err := db.Query(`
		SELECT 
			p.id, p.content, p.image, p.created_at,
			COALESCE((SELECT COUNT(*) FROM likes WHERE post_id = p.id), 0) AS likes_count,
			EXISTS (SELECT 1 FROM likes WHERE post_id = p.id AND user_id = $1) AS liked_by_user
		FROM posts p
		WHERE p.user_id = $2
		ORDER BY p.created_at DESC
	`, userID, profileUserID)

	if err != nil {
		log.Println("error fetching posts: ", err)
		return render.RenderTemplate(c, tmpl, "error", "failed to get posts")
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Content, &post.Image, &post.CreatedAt, &post.LikesCount, &post.LikedByUser)
		if err != nil {
			log.Println("error reading posts: ", err)
			return render.RenderTemplate(c, tmpl, "error", "error reading posts")
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
	rows, err := db.Query(`
        SELECT c.id, c.user_id, u.username, c.content 
        FROM comments c
        JOIN users u ON c.user_id = u.id
        WHERE c.post_id = $1
        ORDER BY c.created_at ASC
    `, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		if err := rows.Scan(&comment.ID, &comment.UserID, &comment.Username, &comment.Content); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func CommentOnPost(c echo.Context, db *sql.DB, tmpl *template.Template) error {
	userId, username := profile.GetCurrentUser(c, db) // Get UserID + Username
	postId := c.Param("id")
	content := c.FormValue("content")

	if content == "" {
		return render.RenderTemplate(c, tmpl, "error", "Content cannot be empty")
	}

	_, err := db.Exec("INSERT INTO comments (post_id, user_id, content, created_at) VALUES ($1, $2, $3, $4)", postId, userId, content, time.Now())
	if err != nil {
		log.Println("Error adding comment: ", err)
		return render.RenderTemplate(c, tmpl, "error", "Error adding comment")
	}

	newComment := struct {
		Username string
		Content  string
	}{
		Username: username,
		Content:  content,
	}

	return render.RenderTemplate(c, tmpl, "single_comment", newComment)
}

func LikePost(c echo.Context, db *sql.DB, tmpl *template.Template) error {
	userID, _ := profile.GetCurrentUser(c, db)
	postID := c.Param("id")

	_, err := db.Exec("INSERT INTO likes (post_id, user_id, created_at) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING", postID, userID, time.Now())
	if err != nil {
		log.Println("Error liking post:", err)
		return render.RenderTemplate(c, tmpl, "error", "Error liking post")
	}

	var likesCount int
	var likedByUser bool
	db.QueryRow("SELECT COUNT(*) FROM likes WHERE post_id = $1", postID).Scan(&likesCount)
	db.QueryRow("SELECT EXISTS(SELECT 1 FROM likes WHERE post_id = $1 AND user_id = $2)", postID, userID).Scan(&likedByUser)

	data := struct {
		ID          string
		LikedByUser bool
		LikesCount  int
	}{
		ID:          postID,
		LikedByUser: likedByUser,
		LikesCount:  likesCount,
	}

	return render.RenderTemplate(c, tmpl, "post-actions", data)
}

func UnlikePost(c echo.Context, db *sql.DB, tmpl *template.Template) error {
	userID, _ := profile.GetCurrentUser(c, db)
	postID := c.Param("id")

	_, err := db.Exec("DELETE FROM likes WHERE post_id = $1 AND user_id = $2", postID, userID)
	if err != nil {
		log.Println("Error unliking post:", err)
		return render.RenderTemplate(c, tmpl, "error", "Error unliking post")
	}

	var likesCount int
	var likedByUser bool
	db.QueryRow("SELECT COUNT(*) FROM likes WHERE post_id = $1", postID).Scan(&likesCount)
	db.QueryRow("SELECT EXISTS(SELECT 1 FROM likes WHERE post_id = $1 AND user_id = $2)", postID, userID).Scan(&likedByUser)

	data := struct {
		ID          string
		LikedByUser bool
		LikesCount  int
	}{
		ID:          postID,
		LikedByUser: likedByUser,
		LikesCount:  likesCount,
	}

	return render.RenderTemplate(c, tmpl, "post-actions", data)
}
