package blober

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// Handler ..
type Handler struct {
	db        *sql.DB
	tablename string
}

func handleGetImage(w http.ResponseWriter, r *http.Request, h Handler) {
	row := h.db.QueryRow("SELECT * FROM "+h.tablename+" WHERE name = $1", r.URL.Path)

	var name string
	var imageBytes []byte

	err := row.Scan(
		&name,
		&imageBytes,
	)

	if err != nil {
		w.WriteHeader(404)
		fmt.Fprintf(w, "404 page not found")
		return
	}

	buffer := bytes.NewBuffer(imageBytes)

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	if _, err := w.Write(buffer.Bytes()); err != nil {
		log.Println("unable to write image.")
	}
}

// Store ..
func (h Handler) Store(name string, b []byte) {
	h.db.Exec("INSERT INTO "+h.tablename+" VALUES($1, $2)", name, b)
}

// Delete removes row with name from the db
func (h Handler) Delete(name string) {
	h.db.Exec("DELETE FROM "+h.tablename+" WHERE name = $1", name)
}

// ServeHTTP implements the http.Handler interface
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handleGetImage(w, r, h)
}

// New returns a handler that serves http requests
// using the db connection and tablename.
func New(db *sql.DB, tablename string) Handler {
	return Handler{db, tablename}
}
