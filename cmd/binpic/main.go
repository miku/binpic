package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
)

// Version is the current version of this tool.
const Version = "0.2.0"

var (
	decode  = flag.Bool("d", false, "decode a binpic-ed png (XXX: not yet implemented)")
	dims    = flag.String("resize", "0x0", "resize, if set")
	output  = flag.String("o", "output.png", "output file, will be a PNG")
	version = flag.Bool("version", false, "show version")
)

// parseDims parses dimensions (like 200x100) or returns 0, if there was any error while parsing.
func parseDims(s string) (width, height int) {
	parts := strings.Split(s, "x")
	if len(parts) != 2 {
		return 0, 0
	}
	var err error
	if width, err = strconv.Atoi(strings.TrimSpace(parts[0])); err != nil {
		return 0, 0
	}
	if height, err = strconv.Atoi(strings.TrimSpace(parts[1])); err != nil {
		return 0, 0
	}
	return width, height
}

// dimsFromSize returns suggested image dimensions given the number of pixels
// e.g. from filesize. The pct parameter can be used to control the ratio,
// e.g. given 0.15 the image height will be 15% less than the square.
func dimsFromSize(size int64, pct float64) (width, height int) {
	sizef := float64(size)
	sq := math.Sqrt(sizef)
	h := math.Ceil(sq - sq*pct)
	w := math.Ceil(sizef / h)
	return int(w), int(h)
}

// Encoder can encode bytes into an image, with optional resize.
type Encoder struct {
	Resize struct {
		W int
		H int
	}
	RatioPct float64
	Fill     uint8
}

// NewEncoder creates an file-to-image encoder with defaults.
func NewEncoder() *Encoder {
	return &Encoder{RatioPct: 0.15, Fill: 255}
}

// shouldResize indicated, whether image should be resized in the process.
func (enc *Encoder) shouldResize() bool {
	return enc.Resize.W > 0 && enc.Resize.H > 0
}

// Encode reads bytes from reader and writes a PNG image to the writer.
// Although a reader is accepted, the implementation might be limited to more
// concrete types. XXX: Accept arbitrary readers through tempfile.
func (enc *Encoder) Encode(w io.Writer, r io.Reader) error {
	f, ok := r.(*os.File)
	if !ok || f == os.Stdin {
		tf, err := ioutil.TempFile("", "binpic-temp-*.file")
		if err != nil {
			return err
		}
		defer os.Remove(tf.Name())
		defer tf.Close()
		if _, err := io.Copy(tf, r); err != nil {
			return err
		}
		if _, err := tf.Seek(0, io.SeekStart); err != nil {
			return err
		}
		f = tf
	}

	fi, err := os.Stat(f.Name())
	if err != nil {
		return err
	}
	width, height := dimsFromSize(fi.Size(), enc.RatioPct)

	// A Rectangle contains the points with Min.X <= X < Max.X, Min.Y <= Y <
	// Max.Y. It is well-formed if Min.X <= Max.X and likewise for Y. Points
	// are always well-formed. A rectangle's methods always return well-formed
	// outputs for well-formed inputs.
	rect := image.Rectangle{
		Min: image.Point{X: 0, Y: 0},          // up left
		Max: image.Point{X: width, Y: height}, // down right
	}
	img := image.NewGray(rect)

	// Reader that allows to read byte per byte.
	br := bufio.NewReader(f)

	// Create a line by line image.
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			b, err := br.ReadByte()
			if err == io.EOF {
				// Fill excess pixels.
				img.Set(x, y, color.Gray{enc.Fill})
				continue
			}
			if err != nil {
				return err
			}
			img.Set(x, y, color.Gray{b})
		}
	}

	var output image.Image = img
	if enc.shouldResize() {
		output = imaging.Resize(img, enc.Resize.W, enc.Resize.H, imaging.Lanczos)
	}
	return png.Encode(w, output)
}

func main() {
	flag.Parse()

	if *version {
		fmt.Printf("%s\n", Version)
		os.Exit(0)
	}
	var r io.Reader = os.Stdin

	if flag.NArg() > 0 {
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		r = f
	}

	of, err := os.Create(*output)
	if err != nil {
		log.Fatal(err)
	}
	defer of.Close()

	bw := bufio.NewWriter(of)
	defer bw.Flush()

	enc := NewEncoder()
	enc.Resize.W, enc.Resize.H = parseDims(*dims)

	if err := enc.Encode(bw, r); err != nil {
		log.Fatal(err)
	}
}
