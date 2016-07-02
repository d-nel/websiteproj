package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/disintegration/imaging"
)

// ImageSaver ..
type ImageSaver interface {
	http.Handler
	Save(img image.Image, filename string)
	Remove(filename string)
}

// ResizeFill ..
func ResizeFill(w int, h int, img image.Image) image.Image {
	return imaging.Fill(img, w, h, imaging.Center, imaging.Lanczos)
}

// ResizeFit ..
func ResizeFit(w int, h int, img image.Image) image.Image {
	return imaging.Fit(img, w, h, imaging.Lanczos)
}

func handleUpload(w http.ResponseWriter, r *http.Request) (image.Image, error) {
	r.ParseMultipartForm(32 << 20)
	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}

type dbSaver struct {
	db    *sql.DB
	table string
}

func (saver dbSaver) Save(img image.Image, filename string) {
	var b bytes.Buffer
	jpeg.Encode(&b, img, nil)

	saver.db.Exec("INSERT INTO "+saver.table+" VALUES($1, $2)", filename, b.Bytes())
}

func (saver dbSaver) Remove(filename string) {
	saver.db.Exec("DELETE FROM "+saver.table+" WHERE name = $1", filename)
}

func (saver dbSaver) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	row := saver.db.QueryRow("SELECT * FROM "+saver.table+" WHERE name = $1", r.URL.Path)

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

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(imageBytes)))
	if _, err := w.Write(imageBytes); err != nil {
		log.Println("unable to write image.")
	}
}

type fsSaver struct {
	handler http.Handler
	path    string
}

func (saver fsSaver) Save(img image.Image, filename string) {
	f, err := os.OpenFile(
		saver.path+filename,
		os.O_WRONLY|os.O_CREATE,
		0666,
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	jpeg.Encode(f, img, nil)
}

func (saver fsSaver) Remove(filename string) {
	os.Remove(saver.path + filename)
}

func (saver fsSaver) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	saver.handler.ServeHTTP(w, r)
}
