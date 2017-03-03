package models

import (
	"fmt"
    "strings"
)

type goPosts struct {
	posts map[string]*Post
}

func GoPosts() *goPosts {
	posts := make(map[string]*Post)
	return &goPosts{posts}
}

func (posts *goPosts) Store(post *Post) error {
	if _, ok := posts.posts[post.ID]; ok {
		return fmt.Errorf("post with id %s already exists", post.ID)
	}
	post.Replies = make(map[string][]string)

	posts.posts[post.ID] = post
	return nil
}

func (posts *goPosts) Update(post *Post) error {
	if _, ok := posts.posts[post.ID]; !ok {
		return fmt.Errorf("can't update post with id %s because it does not exist", post.ID)
	}

	posts.posts[post.ID] = post
	return nil
}

func (posts *goPosts) Delete(id string) error {
	delete(posts.posts, id)
	return nil
}

func (posts *goPosts) ByID(id string) (*Post, error) {
	if _, ok := posts.posts[id]; !ok {
		return nil, fmt.Errorf("damn todo")
	}

	return posts.posts[id], nil
}

func (posts *goPosts) ByUser(uid string) ([]*Post, error) {
	got := make([]*Post, 0)
	for _, post := range posts.posts {
		if post.PostedByID == uid {
			got = append(got, post)
		}
	}

	return got, nil
}

type goUsers struct {
	users map[string]*User
}

func GoUsers() *goUsers {
	users := make(map[string]*User)
	return &goUsers{users}
}

func (users *goUsers) Store(user *User) error {
	if _, ok := users.users[user.ID]; ok {
		return fmt.Errorf("user with id %s already exists", user.ID)
	}

	users.users[user.ID] = user
	return nil
}

func (users *goUsers) Update(user *User) error {
	if _, ok := users.users[user.ID]; !ok {
		return fmt.Errorf("can't update user with id %s because it does not exist", user.ID)
	}

	users.users[user.ID] = user
	return nil
}

func (users *goUsers) Delete(id string) error {
	delete(users.users, id)
	return nil
}

func (users *goUsers) ByID(id string) (*User, error) {
	if _, ok := users.users[id]; !ok {
		return nil, fmt.Errorf("damn todo")
	}

	return users.users[id], nil
}

func (users *goUsers) ByUsername(username string) (*User, error) {
	for _, user := range users.users {
		if user.Username == strings.ToLower(username) {
			return user, nil
		}
	}

	return nil, fmt.Errorf("user by username \"%s\" not found", username)
}

type goSessions struct {
	sessions map[string]*Session
}

func GoSessions() *goSessions {
	sessions := make(map[string]*Session)
	return &goSessions{sessions}
}

func (s *goSessions) GetSession(sid string) (*Session, error) {
	if _, ok := s.sessions[sid]; !ok {
		return nil, fmt.Errorf("damn todo")
	}

	return s.sessions[sid], nil
}

func (s *goSessions) Store(sess *Session) error {
	if _, ok := s.sessions[sess.SID]; ok {
		return fmt.Errorf("session with sid %s already exists", sess.SID)
	}

	s.sessions[sess.SID] = sess
	return nil
}

func (s *goSessions) Delete(sid string) error {
	delete(s.sessions, sid)
	return nil
}
