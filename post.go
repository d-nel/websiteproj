package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/d-nel/websiteproj/models"
)

var posts models.Posts

// add expirey (lol) so I can remove useless files from my hd
var tempPosts map[string]map[string]struct{}

// PostList is a list of Posts
type PostList []*models.Post

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

	post, _ := posts.GetPost(pid)

	if post != nil {
		return genPostID()
	}

	return pid
}

func handleCreatePost(w http.ResponseWriter, r *http.Request) (int, error) {
	user, err := GetUserFromRequest(r)
	if err != nil {
		return 500, err
	}

	if r.Method == http.MethodPost {
		img, err := handleUpload(w, r)
		if err != nil {
			return 500, err
		}

		pid := genPostID()

		SaveImage(img, "/posts/", pid, postSizes[:])

		SaveResizedImageCopy(
			path+"/posts/"+pid+"_preview.jpeg",
			SquareCrop(img),
			256,
		)

		if tempPosts[user.ID] == nil {
			tempPosts[user.ID] = make(map[string]struct{})
		}

		tempPosts[user.ID][pid] = struct{}{}

		data := struct {
			PID string
		}{
			pid,
		}

		js, err := json.Marshal(data)
		if err != nil {
			return 500, err
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}

	return http.StatusOK, nil
}

func handleFinalisePost(w http.ResponseWriter, r *http.Request) (int, error) {
	user, err := GetUserFromRequest(r)
	if err != nil {
		return 500, err
	}

	if r.Method == http.MethodPost {
		pid := r.FormValue("pid")
		replyTo := r.FormValue("replyto")

		post, _ := posts.GetPost(replyTo)

		if post == nil {
			replyTo = ""
		}

		if _, ok := tempPosts[user.ID][pid]; ok {
			RegisterPost(
				pid,
				user.ID,
				replyTo,
			)
		} else {
			//you are a bad person
		}

	}

	return http.StatusOK, nil
}

// RegisterPost ..
func RegisterPost(id string, postedByID string, inReplyTo string) {
	err := posts.Store(
		&models.Post{
			ID:         id,
			PostedByID: postedByID,
			InReplyTo:  inReplyTo,
			PostDate:   time.Now().Unix(),
			ReplyCount: 0,
		},
	)

	if err != nil {
		log.Fatal(err)
	}
}

// GroupPostsHorizontally ..
func GroupPostsHorizontally(postlist []*models.Post, groupSize int) ([][]*models.Post, error) {
	var postsGroup [][]*models.Post
	length := len(postlist)
	for i := 0; i < length; i += groupSize {
		end := i + groupSize

		if end > length {
			end = length
		}

		postsGroup = append(postsGroup, postlist[i:end])
	}

	return postsGroup, nil
}

// SortPostsByDate ..
func SortPostsByDate(postlist []*models.Post) []*models.Post {
	postsByDate := PostList(postlist)
	sort.Sort(sort.Reverse(postsByDate))
	return postsByDate
}
