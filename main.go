package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"

	"time"

	"github.com/d-nel/websiteproj/blober"
	"github.com/d-nel/websiteproj/models"
	_ "github.com/lib/pq"
)

var path string
var blob bool
var prod bool

var tmpl *template.Template

var db *sql.DB

var blobs blober.Handler

// HandleFunc ..
type HandleFunc func(w http.ResponseWriter, r *http.Request) (int, error)

func handleRequest(w http.ResponseWriter, r *http.Request) (int, error) {
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
		return http.StatusNotFound, fmt.Errorf("Page not found")
	}
	return http.StatusOK, nil
}

func handleNewPost(w http.ResponseWriter, r *http.Request) (int, error) {
	handleRefresh(w, r)

	me, _ := GetUserFromRequest(r)
	inReplyTo := r.URL.Query().Get("replyto")

	data := struct {
		Me        *models.User
		InReplyTo string
	}{
		me,
		inReplyTo,
	}

	if err := tmpl.ExecuteTemplate(w, "newpost.html", data); err != nil {
		return 500, err
	}

	return http.StatusOK, nil
}

func handlePostPage(w http.ResponseWriter, r *http.Request) (int, error) {
	handleRefresh(w, r)
	me, _ := GetUserFromRequest(r)

	post, err := posts.GetPost(r.URL.Path[3:])

	if err != nil {
		return http.StatusNotFound, err
	}

	user, _ := users.ByID(post.PostedByID)

	data := struct {
		Me   *models.User
		User *models.User
		Post *models.Post
	}{
		me,
		user,
		post,
	}

	if err := tmpl.ExecuteTemplate(w, "post.html", data); err != nil {
		return 500, err
	}

	return http.StatusOK, nil
}

func handleProfile(w http.ResponseWriter, r *http.Request) (int, error) {
	handleRefresh(w, r)

	me, _ := GetUserFromRequest(r)

	user, err := users.ByUsername(r.URL.Path[3:])
	if err != nil {
		return 404, err
	}

	myProfile := false
	if me != nil {
		myProfile = me.ID == user.ID
		checkTempPosts(me.ID) // brilliant place(!)
	}

	posts, err := posts.GetPostsByUser(user.ID)
	if err != nil {
		return 500, err
	}

	groupedPosts, err := GroupPostsHorizontally(SortPostsByDate(posts), 3)

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

	if err = tmpl.ExecuteTemplate(w, "profile.html", data); err != nil {
		return 500, err
	}

	return http.StatusOK, nil
}

// this function just makes my life easier. it'll be out of here soon.
func handleRefresh(w http.ResponseWriter, r *http.Request) {
	if !prod {
		loadTemplates()
	}
}

// ServeHTTP implements the http.Handler interface
func (fn HandleFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code, err := fn(w, r)

	if err != nil {
		switch code {
		case 404:
			w.WriteHeader(code)
			tmpl.ExecuteTemplate(w, "notfound.html", nil)
		case 500:
			fmt.Println(err)
			w.WriteHeader(code)
			tmpl.ExecuteTemplate(w, "notfound.html", nil)
		default:
			http.Error(w, "ur dumb, m8.", code)
		}
	}
}

func loadTemplates() {
	tmpl = template.New("common")
	tmpl.Funcs(map[string]interface{}{
		"unixformat": timeConverter,
	})
	template.Must(tmpl.ParseGlob(path + "./views/*.html"))
	template.Must(tmpl.ParseGlob(path + "./views/*.tmpl"))
}

func timeConverter(unix int64) string {
	date := time.Unix(unix, 0)
	return date.Format("2 Jan 2006")
}

func staticServe(dir string) {
	fs := http.FileServer(http.Dir(path + "." + dir))
	http.Handle(dir, http.StripPrefix(dir, fs))
}

func serve(patterns map[string]HandleFunc) {
	for pattern, fn := range patterns {
		http.Handle(pattern, fn)
	}
}

func main() {
	fmt.Println("Starting...")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "user=Daniel dbname=userstore sslmode=disable"
	}

	db = models.OpenDB(dbURL)

	path = os.Getenv("RES_PATH")
	if path == "" {
		path = "/Users/Daniel/Dev/Go/src/github.com/d-nel/websiteproj/"
	}

	prod, _ = strconv.ParseBool(os.Getenv("PRODUCTION"))

	users = models.Users{DB: db}
	sessions = models.Sessions{DB: db}
	posts = models.Posts{DB: db}
	tempPosts = make(map[string]map[string]int64)

	loadTemplates()

	staticServe("/static/")

	blobs = blober.New(db, "blobs")
	//http.Handle("/blob/", http.StripPrefix("/blob/", blobs))

	blob, _ = strconv.ParseBool(os.Getenv("BLOB"))
	if blob {
		http.Handle("/data/", http.StripPrefix("/data/", blobs))
		http.Handle("/posts/", http.StripPrefix("/posts/", blobs))
	} else {
		fmt.Println("serving images from file system")
		staticServe("/data/")
		staticServe("/posts/")
	}

	http.HandleFunc("/newpfp", handleProfilePicture)
	http.HandleFunc("/newcover", handleCoverPhoto)

	http.HandleFunc("/register", handleRegister)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/settings", handleSettings)

	serve(map[string]HandleFunc{
		"/newpost": handleNewPost,

		"/delete": handleDeleteConfirm,

		"/post/create":   handleCreatePost,
		"/post/finalise": handleFinalisePost,
		"/post/delete":   handleDeletePost,

		"/user/logout": handleLogout,
		"/user/edit":   handleEditProfile,
		"/user/delete": handleDeleteUser,

		"/":   handleRequest,
		"/p/": handlePostPage,
		"/u/": handleProfile,
	})

	// DONT EVER DO THIS IN PRODUCTION
	http.HandleFunc("/r", handleRefresh)

	var port = os.Getenv("PORT")
	if port == "" {
		port = "8080"
		fmt.Println("defaulting to port: " + port)
	}

	http.ListenAndServe(":"+port, nil)
}
