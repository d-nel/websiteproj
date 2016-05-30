package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/d-nel/websiteproj/models"
)

var users models.Users

// TODO: check db for existing user ids
func genUserID() string {
	b := make([]byte, 8)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data := struct {
			Messsage string
			Username string
		}{
			"",
			"",
		}

		tmpl.ExecuteTemplate(w, "login.html", data)
	} else if r.Method == http.MethodPost {
		username := strings.ToLower(r.FormValue("username"))
		password := r.FormValue("password")

		user, err := users.ByUsername(username)
		ok := false

		if err != nil {
		} else {
			ok = user.CheckPassword(password)
		}

		if ok {
			sid := startSession(user)

			cookie := http.Cookie{Name: "sid", Value: sid, Path: "/", HttpOnly: true}
			http.SetCookie(w, &cookie)

			//TODO rederect new users to welcome page: pfp, desc, etc.
			http.Redirect(w, r, "/", 302)
		} else {
			data := struct {
				Messsage string
				Username string
			}{
				"Incorect username or password",
				username,
			}

			tmpl.ExecuteTemplate(w, "login.html", data)
		}
	}
}

func handleLogout(w http.ResponseWriter, r *http.Request) (int, error) {
	cookie, err := r.Cookie("sid")
	if err != nil {
		return http.StatusBadRequest, err
	}

	sessions.Delete(cookie.Value)

	http.Redirect(w, r, "/", 302)
	return 302, nil
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data := struct {
			Messsage string
			Username string
		}{
			"",
			"",
		}

		tmpl.ExecuteTemplate(w, "register.html", data)
	} else if r.Method == http.MethodPost {
		username := strings.ToLower(r.FormValue("username"))

		user, _ := users.ByUsername(username)

		if user != nil {
			data := struct {
				Messsage string
				Username string
			}{
				"User is already registered",
				"",
			}

			tmpl.ExecuteTemplate(w, "register.html", data)
		} else {
			RegisterUser(
				genUserID(),
				username,
				r.FormValue("password"),
				r.FormValue("email"),
			)

			handleLogin(w, r)
		}
	}
}

func handleEditProfile(w http.ResponseWriter, r *http.Request) (int, error) {
	if r.Method == http.MethodPost {
		user, _ := GetUserFromRequest(r)

		username := r.FormValue("username")
		email := r.FormValue("email")
		name := r.FormValue("name")
		desc := r.FormValue("desc")

		namecheck, _ := users.ByUsername(username)

		if username != "" && namecheck == nil {
			user.Username = strings.ToLower(username)
		}

		if email != "" {
			user.Email = email
		}

		user.Name = name
		user.Description = desc

		err := users.Update(user)

		if err != nil {
			//log.Fatal(err)
			return http.StatusInternalServerError, err
		}

		http.Redirect(w, r, "/u/"+username, 302)
		return 302, err
	}

	return http.StatusMethodNotAllowed, nil
}

func handleDeleteConfirm(w http.ResponseWriter, r *http.Request) (int, error) {
	tmpl.ExecuteTemplate(w, "delete.html", nil)
	return http.StatusOK, nil
}

func handleDeleteUser(w http.ResponseWriter, r *http.Request) (int, error) {
	user, err := GetUserFromRequest(r)
	if err != nil {
		return http.StatusForbidden, nil
	}

	if r.Method == http.MethodPost {
		DeleteUser(user)
		http.Redirect(w, r, "/", 302)
		return http.StatusFound, nil
	}

	return http.StatusOK, nil
}

func handleSettings(w http.ResponseWriter, r *http.Request) {
	handleRefresh(w, r)

	me, err := GetUserFromRequest(r)

	if err != nil {
		fmt.Println(err)
	}

	data := struct {
		Me *models.User
	}{
		me,
	}

	tmpl.ExecuteTemplate(w, "settings.html", data)
}

// RegisterUser ..
func RegisterUser(id string, username string, password string, email string) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	err = users.Store(
		&models.User{
			ID:             id,
			Username:       username,
			HashedPassword: hashedPassword,
			Email:          email,
			Name:           "",
			Description:    "",
		},
	)

	if err != nil {
		log.Fatal(err)
	}
}

// DeleteUser deletes a user and all of their posts
func DeleteUser(user *models.User) {
	posts, _ := posts.ByUser(user.ID)

	for _, post := range posts {
		DeletePost(post)
	}

	deleteUserFiles(user.ID)
	users.Delete(user.ID)
}

// GetUserFromRequest ..
func GetUserFromRequest(r *http.Request) (*models.User, error) {
	cookie, err := r.Cookie("sid")
	if err != nil {
		return nil, err
	}

	sess, err := sessions.GetSession(cookie.Value)
	if err != nil {
		return nil, err
	}

	user, err := users.ByID(sess.UID)
	return user, err
}

func deleteUserFiles(id string) {
	remove := func(name string) {
		os.Remove(path + "/data/" + name)
		blobs.Delete(name)
	}

	remove(id + "_64.jpeg")
	remove(id + "_200.jpeg")
	remove(id + "_480.jpeg")
	remove(id + "_h_1200.jpeg")
}
