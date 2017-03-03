package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
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

var postSizes = [...]int{1024, 512}

// TODO: make db check less memory intense
// TODO: don't forget to check the tmpPosts as well
func genPostID() string {
	b := make([]byte, 8)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	pid := base64.RawURLEncoding.EncodeToString(b)

	post, _ := posts.ByID(pid)
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

		for _, size := range postSizes {
			postImages.Save(
				ResizeFit(size, size, img),
				pid+"_"+strconv.Itoa(size)+".jpeg",
			)
		}

		postImages.Save(
			ResizeFill(256, 256, img),
			pid+"_preview.jpeg",
		)

		if tempPosts[user.ID] == nil {
			tempPosts[user.ID] = make(map[string]int64)
		}

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

func deletePostFiles(pid string) {
	for _, size := range postSizes {
		postImages.Remove(pid + "_" + strconv.Itoa(size) + ".jpeg")
	}

	postImages.Remove(pid + "_preview.jpeg")
}

func handleFinalisePost(w http.ResponseWriter, r *http.Request) (int, error) {
	user, err := GetUserFromRequest(r)
	if err != nil {
		return 500, err
	}

	if r.Method == http.MethodPost {
		pid := r.FormValue("pid")
		replyto := r.FormValue("replyto")

		if _, ok := tempPosts[user.ID][pid]; ok {
			RegisterPost(
				pid,
				user.ID,
			)
			delete(tempPosts[user.ID], pid)

			user.PostCount++
			users.Update(user)
			if replyto != "" {
				re, err := posts.ByID(replyto)
				if err != nil {
					return 500, err
				}

				re.Replies[user.ID] = append(re.Replies[user.ID], pid)
			}

			http.Redirect(w, r, "/u/"+user.Username, 302)
		} else {
			//you are a bad person
		}

	}

	return http.StatusOK, nil
}

func handleDeletePost(w http.ResponseWriter, r *http.Request) (int, error) {
	user, err := GetUserFromRequest(r)
	if err != nil {
		return http.StatusForbidden, nil
	}

	if r.Method == http.MethodPost {
		id := r.FormValue("pid")
		post, _ := posts.ByID(id)

		if post != nil && post.PostedByID == user.ID {

			err := DeletePost(post)

			if err != nil {
				return http.StatusInternalServerError, err
			}

			http.Redirect(w, r, "/u/"+user.Username, 302)
		} else {
			// unauthorized deletion / no such post
			return http.StatusForbidden, nil
		}

	}

	return http.StatusOK, nil
}

// RegisterPost ..
func RegisterPost(id string, postedByID string) {
	err := posts.Store(
		&models.Post{
			ID:         id,
			PostedByID: postedByID,
			PostDate:   time.Now().Unix(),
			ReplyCount: 0,
		},
	)

	if err != nil {
		log.Fatal(err)
	}
}

// DeletePost deletes a post and all of its files
// then adjusts the user's post count
func DeletePost(post *models.Post) error {
	user, _ := users.ByID(post.PostedByID)
	user.PostCount--
	users.Update(user)

	err := posts.Delete(post.ID)
	if err != nil {
		return err
	}

	deletePostFiles(post.ID)

	return nil
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
