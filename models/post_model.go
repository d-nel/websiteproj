package models

import (
	"database/sql"
	"fmt"
)

// Posts deals with the storage of posts aswell as the retrival by certain criteria
type Posts interface {
	Store(post *Post) error
	Update(post *Post) error
	Delete(id string) error

	ByID(id string) (*Post, error)
	ByUser(uid string) ([]*Post, error)
}

// Post is a struct that represents a specific post's infomation in Go
type Post struct {
	ID         string
	PostedByID string
	PostDate   int64
	ReplyCount int
	Replies    map[string][]string // map[By]With
}

type sqlPosts struct {
	*sql.DB
}

// SQLPosts ..
func SQLPosts(db *sql.DB) Posts {
	return &sqlPosts{db}
}

// Store ..
func (posts *sqlPosts) Store(post *Post) error {
	_, err := posts.Exec(
		"INSERT INTO posts VALUES($1, $2, $3, $4)",
		post.ID,
		post.PostedByID,
		post.PostDate,
		post.ReplyCount,
	)

	for by, replies := range post.Replies {
		for _, with := range replies {
			_, err = posts.Exec(
				"INSERT INTO replies VALUES($1, $2, $3)",
				post.ID,
				by,
				with,
			)

			if err != nil {
				return err
			}
		}
	}

	return err
}

//Update ..
func (posts *sqlPosts) Update(post *Post) error {
	_, err := posts.Exec(
		"UPDATE posts SET uid = $2, postdate = $3, replycount = $4 WHERE id = $1",
		post.ID,
		post.PostedByID,
		post.PostDate,
		post.ReplyCount,
	)

	// TODO: this is completly broken
	// it ignores deletion and additions
	for by, replies := range post.Replies {
		for _, with := range replies {
			_, err = posts.Exec(
				"UPDATE replies SET by = $2, with = $3 WHERE toid = $1",
				post.ID,
				by,
				with,
			)

			if err != nil {
				return err
			}
		}
	}

	return err
}

// Delete deletes a post (specified by id) from the db
// TODO: stop ignoring err
func (posts *sqlPosts) Delete(id string) error {
	_, err := posts.Exec(
		"DELETE FROM posts WHERE id = $1",
		id,
	)

	_, err = posts.Exec(
		"DELETE FROM replies WHERE toid = $1",
		id,
	)

	return err
}

// @Bug: multiple replies to the same post can't be saved in one table
// as the id needs to be unique
func (posts *sqlPosts) getReplies(id string) (map[string][]string, error) {
	replies := make(map[string][]string)

	rows, err := posts.Query("SELECT * FROM replies WHERE toid = $1", id)
	if err != nil {
		return replies, err
	}

	for rows.Next() {
		var gotID, by, with string

		err = rows.Scan(
			&gotID,
			&by,
			&with,
		)
		if err != nil {
			return nil, err
		}

		if gotID != id {
			return replies, fmt.Errorf("provided id and replies' id don't match")
		}

		replies[by] = append(replies[by], with)
	}

	return replies, nil
}

// GetPost ..
func (posts *sqlPosts) ByID(id string) (*Post, error) {
	post := new(Post)

	row := posts.QueryRow("SELECT * FROM posts WHERE id = $1", id)
	err := row.Scan(
		&post.ID,
		&post.PostedByID,
		&post.PostDate,
		&post.ReplyCount,
	)
	if err != nil {
		return nil, err
	}

	post.Replies, err = posts.getReplies(id)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (posts *sqlPosts) ByUser(uid string) ([]*Post, error) {
	rows, err := posts.Query("SELECT * FROM posts WHERE uid = $1", uid)
	if err != nil {
		return nil, err
	}

	var postlist []*Post
	for rows.Next() {
		post := new(Post)
		err = rows.Scan(
			&post.ID,
			&post.PostedByID,
			&post.PostDate,
			&post.ReplyCount,
		)
		if err != nil {
			return nil, err
		}

		post.Replies, err = posts.getReplies(post.ID)
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
