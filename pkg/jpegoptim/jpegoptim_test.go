package jpegoptim

// #cgo LDFLAGS: -ljpeg
// #cgo darwin LDFLAGS: -L/opt/local/lib
// #cgo darwin CFLAGS: -I/opt/local/include
// #cgo freebsd LDFLAGS: -L/usr/local/lib
// #cgo freebsd CFLAGS: -I/usr/local/include
// #include <stdlib.h>
// extern int optimizeJPEG(unsigned char *inputbuffer, unsigned long inputsize, unsigned char **outputbuffer, unsigned long *outputsize, int quality);
// extern int encodeJPEG(unsigned char *inputbuffer, int width, int height, unsigned char **outputbuffer, unsigned long *outputsize, int quality);

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"io/ioutil"
	"os"
	"testing"
)

func TestJpegOptimFromBuffer(t *testing.T) {
	fi, err := ioutil.ReadFile("test.jpg")
	if err != nil {
		t.Fatal(err)
	}
	imgBytes, err := EncodeBytesOptimized(fi, &Options{100})
	if err != nil {
		t.Fatal(err)
	}
	_, err = jpeg.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		t.Fatal(err)
	}
	gain := (len(fi) - len(imgBytes)) * 100 / len(fi)
	t.Log("input size", len(fi), "output size", len(imgBytes), "gain", len(fi)-len(imgBytes), gain, "%")
	if gain != 35 {
		t.Fatal("Optimization failed")
	}
}

func TestJpegOptimBadBuffer(t *testing.T) {
	b := []byte{'g', 'o', 'l', 'a', 'n', 'g'}
	_, err := EncodeBytesOptimized(b, &Options{100})
	if err == nil {
		t.Fatal("Should be detected as an error")
	}
}

func TestEncodeImageWithJpegOptim(t *testing.T) {
	m := image.NewRGBA(image.Rect(0, 0, 30, 30))
	m.Set(5, 5, color.RGBA{255, 255, 255, 0})
	w := new(bytes.Buffer)
	err := Encode(w, m, &Options{100})
	if err != nil {
		t.Fatal(err)
	}
	b := w.Bytes()
	if len(b) == 0 {
		t.Fatal("error encoding, size is too small")
	}
	t.Log("output size", len(b), "image 30x30")
	// open output file
	fo, err := os.Create("outputcomp.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer fo.Close()

	fo.Write(b)
}

func BenchmarkOptimize(b *testing.B) {
	m := image.NewRGBA(image.Rect(0, 0, 30, 30))
	m.Set(5, 5, color.RGBA{255, 255, 255, 0})
	w := new(bytes.Buffer)
	for i := 0; i < b.N; i++ {
		err := Encode(w, m, &Options{100})
		if err != nil {
			b.Fatal(err)
		}
	}
}