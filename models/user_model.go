package models

import (
	"database/sql"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Users ...
type Users interface {
	Store(user *User) error
	Update(user *User) error
	Delete(id string) error

	ByID(id string) (*User, error)
	ByUsername(username string) (*User, error)
}

// User is a struct that represents a specific user's infomation from the db in Go
type User struct {
	ID             string
	Username       string
	HashedPassword []byte
	Email          string
	Name           sql.NullString
	Description    sql.NullString
	PostCount      int
}

// Users ..
type sqlUsers struct {
	*sql.DB
}

// SQLUsers ..
func SQLUsers(db *sql.DB) Users {
	return &sqlUsers{db}
}

// CheckPassword checks a plain string password against User's hashed password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword(u.HashedPassword, []byte(password))
	return err == nil
}

// ByUsername queries the db for a user with a maching username
// returns nil, err if not found
func (users *sqlUsers) ByUsername(username string) (*User, error) {
	row := users.QueryRow("SELECT * FROM users WHERE username = $1", strings.ToLower(username))

	return scanUser(row)
}

// ByID ..
func (users *sqlUsers) ByID(id string) (*User, error) {
	row := users.QueryRow("SELECT * FROM users WHERE id = $1", id)

	return scanUser(row)
}

// Store stores "user *User" into the database
func (users *sqlUsers) Store(user *User) error {
	_, err := users.Exec(
		"INSERT INTO users VALUES($1, $2, $3, $4, $5, $6, $7)",
		user.ID,
		user.Username,
		user.HashedPassword,
		user.Email,
		user.Name,
		user.Description,
		user.PostCount,
	)

	return err
}

// Update finds user by user.ID and updates it's row with the data provided
// by the "user *User" argument
func (users *sqlUsers) Update(user *User) error {
	_, err := users.Exec(
		"UPDATE users SET username = $2, password = $3, email = $4, name = $5, description = $6, postcount = $7 WHERE id = $1",
		user.ID,
		user.Username,
		user.HashedPassword,
		user.Email,
		user.Name,
		user.Description,
		user.PostCount,
	)

	return err
}

// Delete deletes a user (specified by id) from the db
func (users *sqlUsers) Delete(id string) error {
	_, err := users.Exec(
		"DELETE FROM users WHERE id = $1",
		id,
	)

	return err
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
		&user.PostCount,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}
