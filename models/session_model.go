package models

import "database/sql"

// Sessions ...
type Sessions interface {
	Store(sess *Session) error
	Delete(sid string) error

	GetSession(sid string) (*Session, error)
}

// Session is a struct that represents Session data from the db
type Session struct {
	SID string
	UID string
}

type sqlSessions struct {
	*sql.DB
}

// SQLSessions ..
func SQLSessions(db *sql.DB) Sessions {
	return &sqlSessions{db}
}

func (sessions *sqlSessions) GetSession(sid string) (*Session, error) {
	row := sessions.QueryRow("SELECT * FROM sessions WHERE sid = $1", sid)

	sess := new(Session)
	err := row.Scan(&sess.SID, &sess.UID)

	return sess, err
}

func (sessions *sqlSessions) Store(sess *Session) error {
	_, err := sessions.Exec(
		"INSERT INTO sessions VALUES($1, $2)",
		sess.SID,
		sess.UID,
	)

	return err
}

func (sessions *sqlSessions) Delete(sid string) error {
	_, err := sessions.Exec(
		"DELETE FROM sessions WHERE sid = $1",
		sid,
	)

	return err
}
