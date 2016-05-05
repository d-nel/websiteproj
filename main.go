package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/d-nel/websiteproj/models"
	_ "github.com/lib/pq"
)

var path string

var tmpl *template.Template

var db *sql.DB

// POST / GET ...
const (
	POST = "POST"
	GET  = "GET"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
	handleRefresh(w, r)
	if r.URL.Path == "/" {
		user, err := GetUserFromRequest(r)

		if err != nil {
			handleLogin(w, r)
		} else {
			data := struct {
				Me *models.User
			}{
				user,
			}

			tmpl.ExecuteTemplate(w, "index.html", data)
		}

	} else {
		tmpl.ExecuteTemplate(w, "notfound.html", nil)
	}

}

func handlePostPage(w http.ResponseWriter, r *http.Request) {
	handleRefresh(w, r)
	me, _ := GetUserFromRequest(r)

	post, err := posts.GetPost(r.URL.Path[3:])

	if err != nil {
		tmpl.ExecuteTemplate(w, "notfound.html", nil)
		return
	}
	user, _ := users.GetUser(post.PostedByID)

	data := struct {
		Me   *models.User
		User *models.User
		Post *models.Post
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
		tmpl.ExecuteTemplate(w, "notfound.html", nil)
		return
	}

	myProfile := false
	if me != nil {
		myProfile = me.ID == user.ID
	}

	posts, err := posts.GetPostsByUser(user.ID)
	groupedPosts, err := GroupPostsHorizontally(SortPostsByDate(posts), 3)

	if err != nil {
		fmt.Println(err)
	}

	data := struct {
		Me           *models.User
		User         *models.User
		GroupedPosts [][]*models.Post
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

// this function just makes my life easier. it'll be out of here soon.
func handleRefresh(w http.ResponseWriter, r *http.Request) {
	loadTemplates()
}

func loadTemplates() {
	tmpl = template.New("common")
	tmpl.ParseGlob(path + "./views/*.html")
	tmpl.ParseGlob(path + "./views/*.tmpl")
}

func staticServe(dir string) {
	fs := http.FileServer(http.Dir(path + dir))
	http.Handle(dir, http.StripPrefix(dir, fs))
}

func main() {
	//"user=Daniel dbname=userstore sslmode=disable"
	fmt.Println("Starting...")
	db = models.OpenDB(os.Getenv("DATABASE_URL"))
	path = os.Getenv("RES_PATH")

	users = models.Users{DB: db}
	sessions = models.Sessions{DB: db}
	posts = models.Posts{DB: db}

	loadTemplates()

	staticServe("./data/")
	staticServe("./posts/")
	staticServe("./static/")

	http.HandleFunc("/newpost", handleCreatePost)
	http.HandleFunc("/newpfp", handleProfilePicture)
	http.HandleFunc("/newcover", handleCoverPhoto)

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

	var port = os.Getenv("PORT")
	if port == "" {
		port = "8080"
		fmt.Println("defaulting to port: " + port)
	}

	http.ListenAndServe(":"+port, nil)
}
