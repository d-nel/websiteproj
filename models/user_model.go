package models

import (
	"database/sql"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// UserStore ...
type UserStore interface {
	Store(user *User) error
	Update(user *User) error
	//Delete(id string) error

	GetUser(id string) (*User, error)
	GetUserByUsername(username string) (*User, error)
}

// User is a struct that represents a specific user's infomation from the db in Go
type User struct {
	ID             string
	Username       string
	HashedPassword []byte
	Email          string
	Name           sql.NullString
	Description    sql.NullString
}

// Users ..
type Users struct {
	DB *sql.DB
}

// CheckPassword checks a plain string password against User's hashed password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword(u.HashedPassword, []byte(password))
	return err == nil
}

// GetUserByUsername queries the db for a user with a maching username
// returns nil, err if not found
func (users *Users) GetUserByUsername(username string) (*User, error) {
	row := users.DB.QueryRow("SELECT * FROM users WHERE username = $1", strings.ToLower(username))

	return scanUser(row)
}

// GetUser ..
func (users *Users) GetUser(id string) (*User, error) {
	row := users.DB.QueryRow("SELECT * FROM users WHERE id = $1", id)

	return scanUser(row)
}

// Store stores "user *User" into the database
func (users *Users) Store(user *User) error {
	_, err := users.DB.Exec(
		"INSERT INTO users VALUES($1, $2, $3, $4, $5, $6)",
		user.ID,
		user.Username,
		user.HashedPassword,
		user.Email,
		user.Name,
		user.Description,
	)

	return err
}

// Update finds user by user.ID and updates it's row with the data provided
// by the "user *User" argument
func (users *Users) Update(user *User) error {
	_, err := users.DB.Exec(
		"UPDATE users SET username = $2, password = $3, email = $4, name = $5, description = $6 WHERE id = $1",
		user.ID,
		user.Username,
		user.HashedPassword,
		user.Email,
		user.Name,
		user.Description,
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
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}