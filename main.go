package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	"github.com/d-nel/websiteproj/models"
	_ "github.com/lib/pq"
)

const path = "/Users/Daniel/Dev/Go/src/github.com/d-nel/websiteproj"

var tmpl *template.Template

var db *sql.DB

// POST ...
const POST = "POST"

// GET ...
const GET = "GET"

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
	user, _ := users.GetUser(post.PostedByID)

	data := struct {
		Me   *models.User
		User *models.User
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

	user, err := users.GetUserByUsername(r.URL.Path[3:])
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
		Me           *models.User
		User         *models.User
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

func handleRefresh(w http.ResponseWriter, r *http.Request) {
	tmpl = template.New("common")
	tmpl.ParseGlob(path + "/views/*.html")
	tmpl.ParseGlob(path + "/views/*.tmpl")
}

func main() {
	db = models.OpenDB("user=Daniel dbname=userstore sslmode=disable")

	users = models.Users{DB: db}
	sessions = models.Sessions{DB: db}

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
	http.HandleFunc("/logout", handleLogout)
	http.HandleFunc("/editprofile", handleEditProfile)
	http.HandleFunc("/settings", handleSettings)

	http.HandleFunc("/", handleRequest)
	http.HandleFunc("/p/", handlePostPage)
	http.HandleFunc("/u/", handleProfile)

	// DONT EVER DO THIS IN PRODUCTION
	http.HandleFunc("/r", handleRefresh)

	http.ListenAndServe(":8080", nil)
}
