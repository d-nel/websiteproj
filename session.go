package main

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/url"

	"github.com/d-nel/websiteproj/models"
)

var sessions models.Sessions

// TODO: check db for existing sessions
func genSessionID() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return url.QueryEscape(base64.URLEncoding.EncodeToString(b))
}

func startSession(u *models.User) string {
	sid := genSessionID()

	_ = sessions.Store(&models.Session{SID: sid, UID: u.ID})

	return sid
}
