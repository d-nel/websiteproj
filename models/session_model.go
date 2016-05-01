package models

import "database/sql"

// SessionStore ...
type SessionStore interface {
	Store(sess *Session) error
	Delete(sid string) error

	GetSession(sid string) (*Session, error)
}

// Session is a struct that represents Session data from the db
type Session struct {
	SID string
	UID string
}

// Sessions ..
type Sessions struct {
	DB *sql.DB
}

// GetSession ..
func (sessions *Sessions) GetSession(sid string) (*Session, error) {
	row := sessions.DB.QueryRow("SELECT * FROM sessions WHERE sid = $1", sid)

	sess := new(Session)
	err := row.Scan(&sess.SID, &sess.UID)

	return sess, err
}

// Store ..
func (sessions *Sessions) Store(sess *Session) error {
	_, err := sessions.DB.Exec(
		"INSERT INTO sessions VALUES($1, $2)",
		sess.SID,
		sess.UID,
	)

	return err
}

// Delete ..
func (sessions *Sessions) Delete(sid string) error {
	_, err := sessions.DB.Exec(
		"DELETE FROM sessions WHERE sid = $1",
		sid,
	)

	return err
}
