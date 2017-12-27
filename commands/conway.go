// An implementation of Conway's Game of Life.
package commands

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"math/rand"
)

// Field represents a two-dimensional field of cells.
type Field struct {
	s    [][]bool
	w, h int
}

// NewField returns an empty field of the specified width and height.
func NewField(w, h int) *Field {
	s := make([][]bool, h)
	for i := range s {
		s[i] = make([]bool, w)
	}
	return &Field{s: s, w: w, h: h}
}

// Set sets the state of the specified cell to the given value.
func (f *Field) Set(x, y int, b bool) {
	f.s[y][x] = b
}

// Alive reports whether the specified cell is alive.
// If the x or y coordinates are outside the field boundaries they are wrapped
// toroidally. For instance, an x value of -1 is treated as width-1.
func (f *Field) Alive(x, y int) bool {
	x += f.w
	x %= f.w
	y += f.h
	y %= f.h
	return f.s[y][x]
}

// Next returns the state of the specified cell at the next time step.
func (f *Field) Next(x, y int) bool {
	// Count the adjacent cells that are alive.
	alive := 0
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if (j != 0 || i != 0) && f.Alive(x+i, y+j) {
				alive++
			}
		}
	}
	// Return next state according to the game rules:
	//   exactly 3 neighbors: on,
	//   exactly 2 neighbors: maintain current state,
	//   otherwise: off.
	return alive == 3 || alive == 2 && f.Alive(x, y)
}

// Life stores the state of a round of Conway's Game of Life.
type Life struct {
	a, b *Field
	w, h int
}

// NewLife returns a new Life game state with a random initial state.
func NewLife(w, h int) *Life {
	a := NewField(w, h)
	for i := 0; i < (w * h / 4); i++ {
		a.Set(rand.Intn(w), rand.Intn(h), true)
	}
	return &Life{
		a: a, b: NewField(w, h),
		w: w, h: h,
	}
}

// Step advances the game by one instant, recomputing and updating all cells.
func (l *Life) Step() {
	// Update the state of the next field (b) from the current field (a).
	for y := 0; y < l.h; y++ {
		for x := 0; x < l.w; x++ {
			l.b.Set(x, y, l.a.Next(x, y))
		}
	}
	// Swap fields a and b.
	l.a, l.b = l.b, l.a
}

type frame struct {
	i   int
	img *image.Paletted
}

func conway(ctx *Context, raw string, args ...string) {
	var w, h, steps = 200, 200, 100

	l := NewLife(w, h)

	frames := make([]*image.Paletted, steps)

	rect := image.Rect(0, 0, w, h)
	pal := color.Palette{color.RGBA{0, 0, 0, 255}, color.RGBA{255, 255, 255, 255}}

	delay := make([]int, steps)
	ch := make(chan frame)

	for i := 0; i < steps; i++ {
		delay[i] = 5

		go func(i int) {
			f := image.NewPaletted(rect, pal)

			for y, row := range l.b.s {
				for x, alive := range row {
					if alive {
						f.SetColorIndex(y, x, 1)
					} else {
						f.SetColorIndex(y, x, 0)
					}
				}
			}

			ch <- frame{i, f}
		}(i)

		l.Step()
	}

	for i := 0; i < steps; i++ {
		f := <-ch
		frames[f.i] = f.img
	}

	gi := &gif.GIF{
		Image: frames,
		Delay: delay,
		Config: image.Config{
			ColorModel: pal,
			Width:      w,
			Height:     h,
		},
	}

	buf := new(bytes.Buffer)

	gif.EncodeAll(buf, gi)

	ctx.ChannelFileSend(ctx.ChannelID, "gol.gif", buf)
}
