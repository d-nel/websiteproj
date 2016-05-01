package models

import (
	"database/sql"
	"time"
)

// PostStore ...
type PostStore interface {
	Store(post *Post) error

	GetPost(id string) (*Post, error)
}

// Post is a struct that represents a specific post's infomation in Go
type Post struct {
	ID         string
	PostedByID string
	InReplyTo  string
	PostDate   int64
	ReplyCount int
}

// Posts ..
type Posts struct {
	DB *sql.DB
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
func (posts *Posts) GetPost(id string) (*Post, error) {
	row := posts.DB.QueryRow("SELECT * FROM posts WHERE id = $1", id)

	return scanPost(row)
}

// GetPostsByUser queries the db for all posts by a user
func (posts *Posts) GetPostsByUser(uid string) ([]*Post, error) {
	rows, err := posts.DB.Query("SELECT * FROM posts WHERE uid = $1", uid)
	if err != nil {
		return nil, err
	}

	var postlist []*Post
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
		postlist = append(postlist, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return postlist, nil
}

// Store ..
func (posts *Posts) Store(post *Post) error {
	_, err := posts.DB.Exec(
		"INSERT INTO posts VALUES($1, $2, $3, $4, $5)",
		post.ID,
		post.PostedByID,
		post.InReplyTo,
		time.Now().Unix(),
		0,
	)

	return err
}
