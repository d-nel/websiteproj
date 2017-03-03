package main

import (
	"image"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestFSSaver(t *testing.T) {
	dir := os.TempDir() + "d-nel/"
	os.Mkdir(dir, os.ModeDir|os.ModePerm)

	saver := fsSaver{
		http.FileServer(http.Dir(dir)),
		dir,
	}

	ts := httptest.NewServer(saver)
	defer ts.Close()

	img := image.NewNRGBA(image.Rect(0, 0, 1000, 1000))
	saver.Save(img, "test.jpg")

	if _, err := os.Stat(dir + "test.jpg"); os.IsNotExist(err) {
		t.Errorf("image does not exist after saving")
		return
	}

	res, err := http.Get(ts.URL + "/test.jpg")
	if err != nil {
		t.Error(err)
		return
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("invaid http status code: %s", res.Status)
		return
	}

	got, _, err := image.Decode(res.Body)
	defer res.Body.Close()

	if err != nil {
		t.Errorf("Served image couldn't be decoded: %s", err)
		return
	}

	if got.Bounds() != img.Bounds() {
		t.Error("Orginal and served image aren't the same size")
		return
	}

	saver.Remove("test.jpg")

	if _, err := os.Stat(dir + "test.jpg"); err == nil {
		t.Error("Couldn't remove image")
		return
	}
}

func TestResizeFill(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 256, 512))
	got := ResizeFill(1000, 1000, img)

	if got.Bounds().Max.X != 1000 || got.Bounds().Max.Y != 1000 {
		t.Errorf("expected 1000x1000 got %dx%d", got.Bounds().Max.X, got.Bounds().Max.Y)
	}
}

func TestResizeFit(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 1000, 2000))
	got := ResizeFit(800, 800, img)

	if got.Bounds().Max.X != 400 && got.Bounds().Max.Y != 800 {
		t.Errorf("expected 400x800 got %dx%d", got.Bounds().Max.X, got.Bounds().Max.Y)
	}
}
