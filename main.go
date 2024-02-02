package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"math"
	"net/http"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const DEBUG = false

var buf []byte

type Wheel struct {
	w, h, cw, ch, r float64

	rect    image.Rectangle
	palette color.Palette

	items  []string
	colors []color.Color

	currentAngle float64
	accelerate   float64

	images []*image.Paletted
	delays []int
}

type MyEvent struct {
	Name string `json:"name"`
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	itemParam := request.QueryStringParameters["items"]
	if itemParam == "" {
		return nil, fmt.Errorf("Missing item query parameter. Please specify the url with ?items=csv,seperated,string,of,items")
	}

	items := strings.Split(itemParam, ",")
	if len(items) == 0 || len(items) > 20 {
		return nil, fmt.Errorf("Invalid number of items.")
	}

	b := buildGif(items)
	b64 := base64.StdEncoding.EncodeToString(b)

	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type":   "image/gif",
			"Content-Length": fmt.Sprintf("%d", len(b64)),
		},
		Body:            b64,
		IsBase64Encoded: true,
	}, nil

}

func main() {
	lambda.Start(HandleRequest)
}

func NewWheel(frames, w, h, r int, colors, p []color.Color, items []string) *Wheel {
	return &Wheel{
		w:       float64(w),
		h:       float64(h),
		cw:      float64(w / 2),
		ch:      float64(h / 2),
		r:       float64(r),
		rect:    image.Rect(0, 0, w, h),
		palette: p,
		items:   items,
		colors:  colors,

		delays: make([]int, 0, frames),
		images: make([]*image.Paletted, 0, frames),
	}
}

func (w *Wheel) Draw(f, delay int) {
	img := image.NewPaletted(w.rect, w.palette)

	if DEBUG {
		drawLabel(img, 10, 10, fmt.Sprintf("%d", f), w.palette[1])
		fmt.Printf("Frame: %d\n", f)
	}

	delta := (2 * math.Pi) / float64(len(w.items))
	for item_i, item := range w.items {
		p1 := Point{x: int(w.cw), y: int(w.ch)}

		c := math.Cos(w.currentAngle)
		s := math.Sin(w.currentAngle)

		p2 := Point{
			x: int(w.cw + w.r*c),
			y: int(w.ch + w.r*s),
		}

		c = math.Cos(w.currentAngle + delta)
		s = math.Sin(w.currentAngle + delta)

		p3 := Point{
			x: int(w.cw + w.r*c),
			y: int(w.ch + w.r*s),
		}

		// main triangle
		DrawFilledTriangle(img, &p1, &p2, &p3, w.colors[item_i*2])

		// arc
		cm := math.Cos(w.currentAngle + (delta / 2))
		sm := math.Sin(w.currentAngle + (delta / 2))
		cp := Point{
			x: int(w.cw + (w.r)*cm),
			y: int(w.ch + (w.r)*sm),
		}

		c1 := math.Cos(w.currentAngle + (delta / 4))
		s1 := math.Sin(w.currentAngle + (delta / 4))
		m1 := Point{
			x: int(w.cw + (w.r)*c1),
			y: int(w.ch + (w.r)*s1),
		}

		c2 := math.Cos(w.currentAngle + ((delta / 4) * 3))
		s2 := math.Sin(w.currentAngle + ((delta / 4) * 3))
		m2 := Point{
			x: int(w.cw + (w.r)*c2),
			y: int(w.ch + (w.r)*s2),
		}

		DrawFilledTriangle(img, &p2, &m1, &cp, w.colors[item_i*2])
		DrawFilledTriangle(img, &p2, &cp, &p3, w.colors[item_i*2])
		DrawFilledTriangle(img, &cp, &m2, &p3, w.colors[item_i*2])

		// label
		labelPoint := Point{
			x: int(w.cw + (w.r-50)*cm),
			y: int(w.ch + (w.r-50)*sm),
		}

		drawLabel(img, labelPoint.x, labelPoint.y, item, w.palette[0])

		w.currentAngle += delta
	}

	if f < 30 {
		w.accelerate += 0.025
	} else {
		w.accelerate -= 0.025
	}

	// draw line
	p1 := Point{x: int(w.cw), y: int(w.ch - w.r - 10)}
	for i := 0; i < 30; i++ {
		img.Set(p1.x-1, p1.y+i, w.palette[1])
		img.Set(p1.x, p1.y+i, w.palette[1])
		img.Set(p1.x+1, p1.y+i, w.palette[1])
	}

	w.currentAngle += w.accelerate

	w.images = append(w.images, img)
	w.delays = append(w.delays, delay)
}

//export getPtr
func getPtr(size int) *byte {
	buf = make([]byte, size)
	return &buf[0]
}

//export getLength
func getLength() int {
	return len(buf)
}

// This function is exported to JavaScript, so can be called using
// exports.multiply() in JavaScript.
//
//export buildGif
func buildGif(items []string) []byte {
	globalPalette := []color.RGBA{
		{R: 3, G: 71, B: 50, A: 255},
		{R: 0, G: 129, B: 72, A: 255},
		{R: 198, G: 192, B: 19, A: 255},
		{R: 239, G: 138, B: 23, A: 255},
		{R: 239, G: 41, B: 23, A: 255},
		{R: 6, G: 214, B: 160, A: 255},
	}

	colors := make([]color.Color, 0, len(items)*2)
	for i := range items {
		c := globalPalette[i%len(globalPalette)]
		r, g, b, _ := c.RGBA()
		colors = append(colors, c, color.RGBA{
			uint8((255 - r) % 255),
			uint8((255 - g) % 255),
			uint8((255 - b) % 255),
			255})
	}

	frames := 60

	p := append(
		[]color.Color{
			color.RGBA{255, 255, 255, 255},
			color.RGBA{0, 0, 0, 255},
		},
		colors...,
	)

	wheel := NewWheel(60, 600, 600, 250, colors, p, items)

	wheel.Draw(0, 100)

	for f := 0; f < frames; f++ {
		wheel.Draw(f, 10)
	}

	wheel.Draw(100, 500)

	// f, err := os.OpenFile("rgb.gif", os.O_WRONLY|os.O_CREATE, 0600)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return nil
	// }
	//
	// defer f.Close()

	var b bytes.Buffer
	bb := bufio.NewWriter(&b)

	err := gif.EncodeAll(bb, &gif.GIF{
		Image:           wheel.images,
		Delay:           wheel.delays,
		BackgroundIndex: 0,
	})

	if err != nil {
		fmt.Println(err)
		return nil
	}

	return b.Bytes()
}

func drawLabel(img *image.Paletted, x, y int, label string, col color.Color) {
	point := fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	v := d.MeasureString(label)
	d.Dot.X -= v / 2
	d.DrawString(label)
}

func Interpolate(i0, d0, i1, d1 int) []int {
	if i0 == i1 {
		return []int{d0}
	}

	values := make([]int, 0, i1-i0+1)
	a := (float32(d1) - float32(d0)) / (float32(i1) - float32(i0))
	d := float32(d0)

	for i := i0; i <= i1; i++ {
		values = append(values, int(d))
		d += a
	}

	return values
}

type Point struct {
	x, y int
}

func DrawFilledTriangle(img *image.Paletted, p0, p1, p2 *Point, col color.Color) {
	// Sort the points from bottom to top.
	if p1.y < p0.y {
		swap := p0
		p0 = p1
		p1 = swap
	}
	if p2.y < p0.y {
		swap := p0
		p0 = p2
		p2 = swap
	}
	if p2.y < p1.y {
		swap := p1
		p1 = p2
		p2 = swap
	}

	// Compute X coordinates of the edges.
	x01 := Interpolate(p0.y, p0.x, p1.y, p1.x)
	x12 := Interpolate(p1.y, p1.x, p2.y, p2.x)
	x02 := Interpolate(p0.y, p0.x, p2.y, p2.x)

	// Merge the two short sides.
	x01 = x01[:len(x01)-1]
	x012 := append(x01, x12...)

	// Determine which is left and which is right.
	var x_left, x_right []int
	m := (len(x02) / 2) | 0

	if x02[m] < x012[m] {
		x_left = x02
		x_right = x012
	} else {
		x_left = x012
		x_right = x02
	}

	// Draw horizontal segments.
	c := uint8(img.Palette.Index(col))
	for y := p0.y; y <= p2.y; y++ {
		for x := x_left[y-p0.y]; x <= x_right[y-p0.y]; x++ {
			i := img.PixOffset(x, y)
			img.Pix[i] = c
		}
	}
}

func drawCircle(img draw.Image, x0, y0, r int, c color.Color) {
	x, y, dx, dy := r-1, 0, 1, 1
	err := dx - (r * 2)

	println("OK")
	println("OK")

	for x > y {
		img.Set(x0+x, y0+y, c)
		img.Set(x0+y, y0+x, c)
		img.Set(x0-y, y0+x, c)
		img.Set(x0-x, y0+y, c)
		img.Set(x0-x, y0-y, c)
		img.Set(x0-y, y0-x, c)
		img.Set(x0+y, y0-x, c)
		img.Set(x0+x, y0-y, c)

		if err <= 0 {
			y++
			err += dy
			dy += 2
		}
		if err > 0 {
			x--
			dx += 2
			err += dx - (r * 2)
		}
	}
}
