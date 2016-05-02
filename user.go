package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
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
	if r.Method == GET {
		data := struct {
			Messsage string
			Username string
		}{
			"",
			"",
		}

		tmpl.ExecuteTemplate(w, "login.html", data)
	} else if r.Method == POST {
		username := strings.ToLower(r.FormValue("username"))
		password := r.FormValue("password")

		user, err := users.GetUserByUsername(username)
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

func handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sid")
	if err != nil {
		return
	}

	sessions.Delete(cookie.Value)

	http.Redirect(w, r, "/", 302)
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method == GET {
		data := struct {
			Messsage string
			Username string
		}{
			"",
			"",
		}

		tmpl.ExecuteTemplate(w, "register.html", data)
	} else if r.Method == POST {
		username := strings.ToLower(r.FormValue("username"))

		user, _ := users.GetUserByUsername(username)

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

func handleEditProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method == POST {
		user, _ := GetUserFromRequest(r)

		username := r.FormValue("username")
		email := r.FormValue("email")
		name := r.FormValue("name")
		desc := r.FormValue("desc")

		namecheck, _ := users.GetUserByUsername(username)

		if username == "" || namecheck != nil {
			username = user.Username
		}

		if email == "" {
			email = user.Email
		}

		/* if form has empty name and desc then it's for a reason
		if name == "" {
			name = user.Name.String
		}

		if desc == "" {
			desc = user.Description.String
		}
		*/

		err := users.Update(
			&models.User{
				ID:             user.ID,
				Username:       strings.ToLower(username),
				HashedPassword: user.HashedPassword,
				Email:          email,
				Name:           sql.NullString{String: name, Valid: true},
				Description:    sql.NullString{String: desc, Valid: true},
			},
		)

		if err != nil {
			log.Fatal(err)
		}
	}
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
			Name:           sql.NullString{},
			Description:    sql.NullString{},
		},
	)

	if err != nil {
		log.Fatal(err)
	}
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

	user, err := users.GetUser(sess.UID)
	return user, err
}
