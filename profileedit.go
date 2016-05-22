package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"net/http"
	"os"
	"strconv"

	"github.com/nfnt/resize"
)

var pfpSizes = [...]uint{480, 200, 64}

// upload logic
func handleProfilePicture(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromRequest(r)
	if err != nil {
		fmt.Println(err)
		return
	}

	if r.Method == http.MethodPost {
		img, err := handleUpload(w, r)
		if err != nil {
			fmt.Println(err)
			return
		}

		SaveImage(SquareCrop(img), "/data/", user.ID, pfpSizes[:])
	}

	http.Redirect(w, r, "/", 302)
}

func handleCoverPhoto(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromRequest(r)
	if err != nil {
		fmt.Println(err)
		return
	}

	if r.Method == http.MethodPost {
		img, err := handleUpload(w, r)
		if err != nil {
			fmt.Println(err)
			return
		}

		dx := 1200
		dy := 370

		// scale to match width with dx
		// unless the height is less than dy
		if img.Bounds().Dy() < dy {
			img = resize.Resize(0, uint(dy), img, resize.Lanczos3)
		} else {
			img = resize.Resize(uint(dx), 0, img, resize.Lanczos3)
		}

		crop := image.NewRGBA(image.Rect(0, 0, dx, dy))
		draw.Draw(crop, crop.Bounds(), img, image.ZP, draw.Src)

		SaveImage(crop, "/data/", user.ID+"_h", []uint{1200})
	}

	http.Redirect(w, r, "/", 302)
}

// SaveImage ...
func SaveImage(img image.Image, subpath string, id string, sizes []uint) {
	for size := 0; size < len(sizes); size++ {
		SaveResizedImageCopy(
			path+subpath+id+"_"+strconv.Itoa(int(sizes[size]))+".jpeg",
			img,
			sizes[size],
		)
	}
}

// SquareCrop TODO: right now it crops to the top corner - BAD
func SquareCrop(img image.Image) image.Image {
	dx := img.Bounds().Dx()
	dy := img.Bounds().Dy()

	if dx < dy {
		dy = dx
	} else {
		dx = dy
	}

	crop := image.NewRGBA(image.Rect(0, 0, dx, dy))
	draw.Draw(crop, crop.Bounds(), img, image.ZP, draw.Src)

	return crop
}

// SaveResizedImageCopy ..
func SaveResizedImageCopy(filename string, img image.Image, size uint) {
	var dx uint
	var dy uint

	if img.Bounds().Dx() < img.Bounds().Dy() {
		dy = size
	} else {
		dx = size
	}

	imgResize := resize.Resize(dx, dy, img, resize.Lanczos3)

	f, err := os.OpenFile(
		filename,
		os.O_WRONLY|os.O_CREATE,
		0666,
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	jpeg.Encode(f, imgResize, nil)

	// TODO: full path is saved. change to proper name
	// var b bytes.Buffer
	// jpeg.Encode(&b, imgResize, nil)
	// blobs.SaveBytes(filename, b.Bytes())
}

func handleUpload(w http.ResponseWriter, r *http.Request) (image.Image, error) {
	r.ParseMultipartForm(32 << 20)
	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return img, nil
}
