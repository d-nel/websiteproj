package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"

	"time"

	"github.com/d-nel/websiteproj/models"
	_ "github.com/lib/pq"
)

var path string
var blob bool
var prod bool

var tmpl *template.Template

var db *sql.DB

var dataImages ImageSaver
var postImages ImageSaver

// HandleFunc ..
type HandleFunc func(w http.ResponseWriter, r *http.Request) (int, error)

func handleRequest(w http.ResponseWriter, r *http.Request) (int, error) {
	if r.URL.Path == "/" {
		user, err := GetUserFromRequest(r)

		if err != nil {
			// if user isn't auth'd redirect to login page
			http.Redirect(w, r, "/login", 302)
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
	me, _ := GetUserFromRequest(r)

	post, err := posts.ByID(r.URL.Path[3:])

	if err != nil {
		return http.StatusNotFound, err
	}

	user, _ := users.ByID(post.PostedByID)

	type tempReply struct {
		By           *models.User
		With         *models.Post
		WithPostedBy string
	}

	var replies []tempReply

	for byID, withs := range post.Replies {
		for _, withID := range withs {
			by, _ := users.ByID(byID)
			with, err := posts.ByID(withID)
			if err != nil {
				return 500, err
			}

			originalPoster, _ := users.ByID(with.PostedByID)

			withPostedBy := ""

			if by.Username != originalPoster.Username {
				withPostedBy = originalPoster.Username
			}

			replies = append(replies, tempReply{by, with, withPostedBy})
		}
	}

	data := struct {
		Me      *models.User
		User    *models.User
		Post    *models.Post
		Replies []tempReply
	}{
		me,
		user,
		post,
		replies,
	}

	if err := tmpl.ExecuteTemplate(w, "post.html", data); err != nil {
		return 500, err
	}

	return http.StatusOK, nil
}

func handleProfile(w http.ResponseWriter, r *http.Request) (int, error) {
	me, _ := GetUserFromRequest(r)

	user, err := users.ByUsername(r.URL.Path[3:])
	if err != nil {
		return 404, err
	}

	myProfile := false
	if me != nil {
		myProfile = me.ID == user.ID
	}

	posts, err := posts.ByUser(user.ID)
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
	handleRefresh(w, r)
	code, err := fn(w, r)

	// @TODO sometimes I returned (500, nil) which skips this,
	// and that is not necessarily the best. It causes it to write 200 OK
	// even in the case of an err.

	if err != nil {
		w.WriteHeader(code)
		switch code {
		case 404:
			// w.WriteHeader(code)
			tmpl.ExecuteTemplate(w, "notfound.html", nil)
		case 500:
			fmt.Println(err) // @TODO better logging of errors
			// w.WriteHeader(code)
			tmpl.ExecuteTemplate(w, "notfound.html", nil)
		default:
			// fmt.Printf("Undefined error code (%d): %s\n", code, err)
			// w.WriteHeader(code)
			// http.Error(w, "Server error", 500)
		}
	}
}

func loadTemplates() {
	if prod && tmpl != nil {
		return
	}
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

func serveAllRoutes() {
	patterns := map[string]HandleFunc{
		"/newpost": handleNewPost,

		"/delete": handleDeleteConfirm,

		"/post/create":   handleCreatePost,
		"/post/finalise": handleFinalisePost,
		"/post/delete":   handleDeletePost,

		"/user/logout":    handleLogout,
		"/user/edit":      handleEditProfile,
		"/user/editpfp":   handleEditPFP,
		"/user/editcover": handleEditCover,
		"/user/delete":    handleDeleteUser,

		"/settings": handleSettings,

		"/":   handleRequest,
		"/p/": handlePostPage,
		"/u/": handleProfile,
	}

	for pattern, fn := range patterns {
		http.Handle(pattern, fn)
	}
}

func main() {
	fmt.Println("Starting...")

	dummy, _ := strconv.ParseBool(os.Getenv("DUMMY"))

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "user=Daniel dbname=websiteproj sslmode=disable"
	}

	if !dummy {
		db = models.OpenDB(dbURL)
	}

	path = os.Getenv("RES_PATH")

	prod, _ = strconv.ParseBool(os.Getenv("PRODUCTION"))

	if dummy {
		users = models.GoUsers()
		sessions = models.GoSessions()
		posts = models.GoPosts()
	} else {
		users = models.SQLUsers(db)
		sessions = models.SQLSessions(db)
		posts = models.SQLPosts(db)
	}

	// this whole tempPosts solution feels bad
	tempPosts = make(map[string]map[string]int64)
	go func() {
		c := time.Tick(5 * time.Minute)
		for range c {
			for uid := range tempPosts {
				for key, expiry := range tempPosts[uid] {
					if time.Now().Unix() > expiry {
						delete(tempPosts[uid], key)
						deletePostFiles(key)
					}
				}
			}
		}
	}()

	loadTemplates()

	blob, _ = strconv.ParseBool(os.Getenv("BLOB"))
	if blob {
		dataImages = dbSaver{db, "blobs"}
		postImages = dbSaver{db, "blobs"}
	} else if dummy {
		dataImages = dummySaver{make(map[string][]byte)}
		postImages = dummySaver{make(map[string][]byte)}
	} else {
		fmt.Println("serving images from file system")

		dataImages = fsSaver{
			http.FileServer(http.Dir(path + "./data/")),
			path + "./data/",
		}

		postImages = fsSaver{
			http.FileServer(http.Dir(path + "./data/posts/")),
			path + "./data/posts/",
		}
	}

	http.Handle("/data/", http.StripPrefix("/data/", dataImages))
	http.Handle("/posts/", http.StripPrefix("/posts/", postImages))

	staticServe("/static/")

	http.HandleFunc("/register", handleRegister)
	http.HandleFunc("/login", handleLogin)

	serveAllRoutes()

	// DONT EVER DO THIS IN PRODUCTION
	http.HandleFunc("/r", handleRefresh)

	var port = os.Getenv("PORT")
	if port == "" {
		port = "8080"
		fmt.Println("defaulting to port: " + port)
	}

	http.ListenAndServe(":"+port, nil)
}
