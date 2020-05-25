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

// Version of tool.
const Version = "0.2.0"

var (
	decode   = flag.Bool("d", false, "decode a binpic-ed png (XXX: not yet implemented)")
	hasColor = flag.Bool("color", false, "produce an image with colored pixels")
	dims     = flag.String("resize", "0x0", "resize, if set")
	output   = flag.String("o", "output.png", "output file, will be a PNG")
	version  = flag.Bool("version", false, "show version")
	invert   = flag.Bool("invert", false, "invert color")
)

// parseDims parses dimensions (like 200x100) or returns 0, if there was any error while parsing.
func parseDims(s string) (width, height int) {
	var (
		parts = strings.Split(s, "x")
		err   error
	)
	if len(parts) != 2 {
		return 0, 0
	}
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
	var (
		sizef = float64(size)
		sq    = math.Sqrt(sizef)
		h     = math.Ceil(sq - sq*pct)
		w     = math.Ceil(sizef / h)
	)
	return int(w), int(h)
}

func keepColor(c color.RGBA) color.RGBA {
	return c
}

func invertColor(c color.RGBA) color.RGBA {
	return color.RGBA{255 - c.R, 255 - c.G, 255 - c.B, c.A}
}

func calcGreyShade(b byte) color.RGBA {
	return color.RGBA{b, b, b, 0xff}
}

func calcColor(b byte) color.RGBA {
	return color.RGBA{
		((b & 0o300) >> 6) * 64,
		((b & 0o070) >> 3) * 32,
		(b & 0o007) * 32,
		0xff,
	}
}

// Encoder can encode bytes into an image, with optional resize.
type Encoder struct {
	Resize struct {
		W int
		H int
	}
	RatioPct       float64
	Fill           uint8
	ColorTransform func(color.RGBA) color.RGBA
	ColorFunc      func(b byte) color.RGBA
}

// NewEncoder creates an file-to-image encoder with defaults.
func NewEncoder() *Encoder {
	return &Encoder{RatioPct: 0.15, Fill: 255, ColorFunc: calcGreyShade, ColorTransform: keepColor}
}

// shouldResize indicates, whether image should be resized in the process.
func (enc *Encoder) shouldResize() bool {
	return enc.Resize.W > 0 && enc.Resize.H > 0
}

// Encode reads bytes from reader and writes a PNG image to the writer.
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
	var (
		width, height = dimsFromSize(fi.Size(), enc.RatioPct)
		// A Rectangle contains the points with Min.X <= X < Max.X, Min.Y <= Y <
		// Max.Y. It is well-formed if Min.X <= Max.X and likewise for Y. Points
		// are always well-formed. A rectangle's methods always return well-formed
		// outputs for well-formed inputs.
		rect = image.Rectangle{
			Min: image.Point{X: 0, Y: 0},          // up left
			Max: image.Point{X: width, Y: height}, // down right
		}
		img = image.NewRGBA(rect)
		br  = bufio.NewReader(f)
	)
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
			img.Set(x, y, enc.ColorTransform(enc.ColorFunc(b)))
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
	if *hasColor {
		enc.ColorFunc = calcColor
	}
	if *invert {
		enc.ColorTransform = invertColor
	}
	if err := enc.Encode(bw, r); err != nil {
		log.Fatal(err)
	}
}
