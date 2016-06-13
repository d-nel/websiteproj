package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"net/http"
	"os"

	"github.com/disintegration/imaging"
)

// ResizeFill ..
func ResizeFill(w int, h int, img image.Image) image.Image {
	return imaging.Fill(img, w, h, imaging.Center, imaging.Lanczos)
}

// ResizeFit ..
func ResizeFit(w int, h int, img image.Image) image.Image {
	return imaging.Fit(img, w, h, imaging.Lanczos)
}

// SaveImage ..
func SaveImage(img image.Image, filepath string, filename string) {
	if blob {
		var b bytes.Buffer
		jpeg.Encode(&b, img, nil)
		blobs.Store(filename, b.Bytes())
		return
	}

	f, err := os.OpenFile(
		filepath+filename,
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
