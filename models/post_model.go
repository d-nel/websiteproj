package models

import "database/sql"

// Posts ...
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
	InReplyTo  string
	PostDate   int64
	ReplyCount int
}

// Posts ..
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
		"INSERT INTO posts VALUES($1, $2, $3, $4, $5)",
		post.ID,
		post.PostedByID,
		post.InReplyTo,
		post.PostDate,
		post.ReplyCount,
	)

	return err
}

//Update ..
func (posts *sqlPosts) Update(post *Post) error {
	_, err := posts.Exec(
		"UPDATE posts SET uid = $2, inreplyto = $3, postdate = $4, replycount = $5 WHERE id = $1",
		post.ID,
		post.PostedByID,
		post.InReplyTo,
		post.PostDate,
		post.ReplyCount,
	)

	return err
}

// Delete deletes a post (specified by id) from the db
func (posts *sqlPosts) Delete(id string) error {
	_, err := posts.Exec(
		"DELETE FROM posts WHERE id = $1",
		id,
	)

	return err
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
func (posts *sqlPosts) ByID(id string) (*Post, error) {
	row := posts.QueryRow("SELECT * FROM posts WHERE id = $1", id)

	return scanPost(row)
}

// GetPostsByUser queries the db for all posts by a user
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
