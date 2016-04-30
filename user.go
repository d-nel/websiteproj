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
)

//TODO lowercase username all the time
//TODO the desc in general m8

// User is a struct that represents a specific user's infomation from the db in Go
type User struct {
	ID             string
	Username       string
	HashedPassword []byte
	Email          string
	Name           sql.NullString
	Description    sql.NullString
}

// TODO: check db for existing user ids
func genUserID() string {
	b := make([]byte, 8)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// CheckPassword checks a plain string password against User's hashed password
func (u User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword(u.HashedPassword, []byte(password))
	return err == nil
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

		user, err := GetUserByUsername(username)
		ok := false

		if err != nil {
		} else {
			ok = user.CheckPassword(password)
		}

		if ok {
			sid := user.startSession()

			cookie := http.Cookie{Name: "sid", Value: sid, Path: "/", HttpOnly: true}
			http.SetCookie(w, &cookie)

			//TODO rederect new users to welcome page: pfp, disc, etc.
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

	_, err = db.Exec(
		"DELETE FROM sessions WHERE sid = $1",
		cookie.Value,
	)

	if err != nil {
		fmt.Println(err)
		return
	}

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

		user, _ := GetUserByUsername(username)

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

		email := r.FormValue("email")
		name := r.FormValue("name")
		desc := r.FormValue("desc")

		if email == "" {
			email = user.Email
		}

		if name == "" {
			name = user.Name.String
		}

		if desc == "" {
			desc = user.Description.String
		}

		_, err := db.Exec(
			"UPDATE users SET email = $2, name = $3, description = $4 WHERE id = $1",
			user.ID,
			email,
			name,
			desc,
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
		Me *User
	}{
		me,
	}

	tmpl.ExecuteTemplate(w, "settings.html", data)
}

func scanUser(row *sql.Row) (*User, error) {
	user := new(User)
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.HashedPassword,
		&user.Email,
		&user.Name,
		&user.Description,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByUsername queries the db for a user with a maching username
// returns nil, err if not found
func GetUserByUsername(username string) (*User, error) {
	row := db.QueryRow("SELECT * FROM users WHERE username = $1", strings.ToLower(username))

	return scanUser(row)
}

// GetUser ..
func GetUser(id string) (*User, error) {
	row := db.QueryRow("SELECT * FROM users WHERE id = $1", id)

	return scanUser(row)
}

// GetUserFromRequest ..
func GetUserFromRequest(r *http.Request) (*User, error) {
	cookie, err := r.Cookie("sid")
	if err != nil {
		return nil, err
	}

	sess, err := GetSession(cookie.Value)
	if err != nil {
		return nil, err
	}

	user, err := GetUser(sess.UID)
	return user, err
}

// RegisterUser ..
func RegisterUser(id string, username string, password string, email string) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(
		"INSERT INTO users VALUES($1, $2, $3, $4, $5, $6)",
		id,
		username,
		hashedPassword,
		email,
		"",
		"",
	)
	if err != nil {
		log.Fatal(err)
	}
}
