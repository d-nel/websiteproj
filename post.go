package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"time"
)

// Post is a struct that represents a specific post's infomation from the db in Go
type Post struct {
	ID         string
	PostedByID string
	InReplyTo  string
	PostDate   int
	ReplyCount int
}

// PostList is a list of Posts
type PostList []*Post

func (p PostList) Len() int           { return len(p) }
func (p PostList) Less(i, j int) bool { return p[i].PostDate < p[j].PostDate }
func (p PostList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

var postSizes = [...]uint{1024, 512}

// TODO: make db check less memory intense
func genPostID() string {
	b := make([]byte, 8)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	pid := base64.RawURLEncoding.EncodeToString(b)

	post, _ := GetPost(pid)

	if post != nil {
		return genPostID()
	}

	return pid
}

func handleCreatePost(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromRequest(r)
	if err != nil {
		fmt.Println(err)
		return
	}

	if r.Method == POST {
		img, err := handleUpload(w, r)
		if err != nil {
			fmt.Println(err)
			return
		}

		pid := genPostID()

		SaveImage(img, "/posts/", pid, postSizes[:])

		SaveResizedImageCopy(
			path+"/posts/"+pid+"_preview.jpeg",
			SquareCrop(img),
			256,
		)

		RegisterPost(
			pid,
			user.ID,
			"",
		)
	}
}

// RegisterPost ..
func RegisterPost(id string, postedByID string, inReplyTo string) {
	_, err := db.Exec(
		"INSERT INTO posts VALUES($1, $2, $3, $4, $5)",
		id,
		postedByID,
		inReplyTo,
		time.Now().Unix(),
		0,
	)

	if err != nil {
		log.Fatal(err)
	}
}

// GroupPostsHorizontally ..
func GroupPostsHorizontally(posts []*Post, groupSize int) ([][]*Post, error) {
	var postsGroup [][]*Post
	length := len(posts)
	for i := 0; i < length; i += groupSize {
		end := i + groupSize

		if end > length {
			end = length
		}

		postsGroup = append(postsGroup, posts[i:end])
	}

	return postsGroup, nil
}

func scanPost(row *sql.Row) (*Post, error) {
	post := new(Post)
	err := row.Scan(
		&post.ID,
		&post.PostedByID,
		&post.InReplyTo,
		&post.PostDate,
		&post.ReplyCount,
	)

	if err != nil {
		return nil, err
	}

	return post, nil
}

// GetPost ..
func GetPost(id string) (*Post, error) {
	row := db.QueryRow("SELECT * FROM posts WHERE id = $1", id)

	return scanPost(row)
}

// GetPostsByUser queries the db for all posts by a user
func GetPostsByUser(uid string) ([]*Post, error) {
	rows, err := db.Query("SELECT * FROM posts WHERE uid = $1", uid)
	if err != nil {
		return nil, err
	}

	var posts []*Post
	for rows.Next() {
		post := new(Post)
		err = rows.Scan(
			&post.ID,
			&post.PostedByID,
			&post.InReplyTo,
			&post.PostDate,
			&post.ReplyCount,
		)

		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	postsByDate := PostList(posts)
	sort.Sort(sort.Reverse(postsByDate))

	return postsByDate, nil
}
