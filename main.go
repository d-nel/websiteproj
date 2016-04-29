package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"

	_ "github.com/lib/pq"
)

const path = "/Users/Daniel/Dev/Go/src/github.com/d-nel/test"

// Session is a struct that represents Session data from the db
type Session struct {
	SID string
	UID string
}

var tmpl *template.Template

var db *sql.DB

// POST ...
const POST = "POST"

// GET ...
const GET = "GET"

func (u User) startSession() string {
	sid := genSessionID()

	_, err := db.Exec("INSERT INTO sessions VALUES($1, $2)", sid, u.ID)
	if err != nil {
		log.Fatal(err)
	}

	return sid
}

// TODO: check db for existing sessions
func genSessionID() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return url.QueryEscape(base64.URLEncoding.EncodeToString(b))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	handleRefresh(w, r)
	if r.URL.Path == "/" {
		_, err := GetUserFromRequest(r)

		if err != nil {
			handleLogin(w, r)
		} else {
			if err != nil {
				tmpl.ExecuteTemplate(w, "login.html", nil)
			} else {

				fmt.Fprintf(w, "Main page yay")
			}
		}

	} else {
		fmt.Fprint(w, "404 looser!")
	}

}

func handlePostPage(w http.ResponseWriter, r *http.Request) {
	handleRefresh(w, r)
	me, _ := GetUserFromRequest(r)

	post, err := GetPost(r.URL.Path[3:])

	if err != nil {
		fmt.Fprintf(w, "sry pst not found %s", r.URL.Path[3:])
		return
	}
	user, _ := GetUser(post.PostedByID)

	data := struct {
		Me   *User
		User *User
		Post *Post
	}{
		me,
		user,
		post,
	}

	tmpl.ExecuteTemplate(w, "post.html", data)
}

func handleProfile(w http.ResponseWriter, r *http.Request) {
	handleRefresh(w, r)

	me, _ := GetUserFromRequest(r)

	user, err := GetUserByUsername(r.URL.Path[3:])
	if err != nil {
		fmt.Fprintf(w, "sry profile not found %s", r.URL.Path[3:])
		return
	}

	myProfile := false
	if me != nil {
		myProfile = me.ID == user.ID
	}

	posts, err := GetPostsByUser(user.ID)
	groupedPosts, err := GroupPostsHorizontally(posts, 3)

	if err != nil {
		fmt.Println(err)
	}

	data := struct {
		Me           *User
		User         *User
		GroupedPosts [][]*Post
		MyProfile    bool
	}{
		me,
		user,
		groupedPosts,
		myProfile,
	}

	err = tmpl.ExecuteTemplate(w, "profile.html", data)
	if err != nil {
		fmt.Println(err)
	}
}

// GetSession ..
func GetSession(sid string) (*Session, error) {
	row := db.QueryRow("SELECT * FROM sessions WHERE sid = $1", sid)

	sess := new(Session)
	err := row.Scan(&sess.SID, &sess.UID)

	return sess, err
}

func handleRefresh(w http.ResponseWriter, r *http.Request) {
	tmpl = template.New("common")
	tmpl.ParseGlob(path + "/views/*.html")
	tmpl.ParseGlob(path + "/views/*.tmpl")
}

func main() {
	var err error
	db, err = sql.Open("postgres", "user=Daniel dbname=userstore sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	tmpl = template.New("common")
	tmpl.ParseGlob(path + "/views/*.html")

	fs := http.FileServer(http.Dir(path + "/data"))
	http.Handle("/data/", http.StripPrefix("/data/", fs))

	posts := http.FileServer(http.Dir(path + "/posts"))
	http.Handle("/posts/", http.StripPrefix("/posts/", posts))

	static := http.FileServer(http.Dir(path + "/static"))
	http.Handle("/static/", http.StripPrefix("/static/", static))

	http.HandleFunc("/newpost", handleCreatePost)
	http.HandleFunc("/newpfp", handleProfilePicture)

	http.HandleFunc("/register", handleRegister)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/editprofile", handleEditProfile)

	http.HandleFunc("/", handleRequest)
	http.HandleFunc("/p/", handlePostPage)
	http.HandleFunc("/u/", handleProfile)

	// DONT EVER DO THIS IN PRODUCTION
	http.HandleFunc("/r", handleRefresh)

	http.ListenAndServe(":8080", nil)
}
