package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/d-nel/websiteproj/models"
)

var posts models.Posts

// map[user.ID]map[pid]expiry
var tempPosts map[string]map[string]int64

// PostList is a list of Posts
type PostList []*models.Post

func (p PostList) Len() int           { return len(p) }
func (p PostList) Less(i, j int) bool { return p[i].PostDate < p[j].PostDate }
func (p PostList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

var postSizes = [...]uint{1024, 512}

// TODO: make db check less memory intense
// TODO: don't forget to check the tmpPosts as well
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

		checkTempPosts(user.ID)

		tempPosts[user.ID][pid] = time.Now().Add(time.Duration(2) * time.Hour).Unix()

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

// This is not a permanent solution (obviously)
func checkTempPosts(uid string) {
	if tempPosts[uid] == nil {
		tempPosts[uid] = make(map[string]int64)
	}

	for key, t := range tempPosts[uid] {
		if time.Now().Unix() > t {
			delete(tempPosts[uid], key)

			os.Remove(path + "/posts/" + key + "_512.jpeg")
			os.Remove(path + "/posts/" + key + "_1024.jpeg")
			os.Remove(path + "/posts/" + key + "_preview.jpeg")
		}
	}
}

func handleFinalisePost(w http.ResponseWriter, r *http.Request) (int, error) {
	user, err := GetUserFromRequest(r)
	if err != nil {
		return 500, err
	}

	if r.Method == http.MethodPost {
		pid := r.FormValue("pid")
		replyTo := r.FormValue("replyto")

		if _, ok := tempPosts[user.ID][pid]; ok {
			replyToPost, _ := posts.GetPost(replyTo)

			if replyToPost == nil {
				replyTo = ""
			} else {
				replyToPost.ReplyCount++
				posts.Update(replyToPost)
			}

			RegisterPost(
				pid,
				user.ID,
				replyTo,
			)
			delete(tempPosts[user.ID], pid)

			user.PostCount++
			users.Update(user)

			http.Redirect(w, r, "/u/"+user.Username, 302)
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
