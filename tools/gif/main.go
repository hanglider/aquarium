// Command gif рисует анимацию аквариума для README: та же симуляция,
// символы как в терминале, только пиксельные.
//
//	go run ./tools/gif -o aquarium.gif
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"os"

	"github.com/hanglider/aquarium/internal/sim"
)

// Размеры в пикселях: символ 5×7 рисуется в масштабе 2 в клетке 12×18.
const (
	scale        = 2
	cellW, cellH = 12, 18
	margin       = 8 // от края картинки до рамки и от рамки до аквариума
)

// Индексы в палитре. 0 — прозрачный: им помечаются пиксели, которые
// не изменились с прошлого кадра, — так GIF получается в разы меньше.
const (
	transparent = iota
	background
	border
	food
	fish
	predator
)

var palette = color.Palette{
	transparent: color.RGBA{},
	background:  color.RGBA{0x0b, 0x14, 0x16, 0xff},
	border:      color.RGBA{0x1f, 0x3a, 0x3f, 0xff},
	food:        color.RGBA{0x4f, 0xbf, 0x7f, 0xff},
	fish:        color.RGBA{0xe2, 0xb3, 0x3c, 0xff},
	predator:    color.RGBA{0xef, 0x5f, 0x4f, 0xff},
}

// Символы 5×7, как в терминальном шрифте.
var glyphs = map[uint8][7]string{
	food: {
		".....",
		"..#..",
		"#.#.#",
		".###.",
		"#.#.#",
		"..#..",
		".....",
	},
	fish: {
		"#....",
		".#...",
		"..#..",
		"...#.",
		"..#..",
		".#...",
		"#....",
	},
	predator: {
		".###.",
		"#...#",
		"#.###",
		"#.#.#",
		"#.###",
		"#....",
		".###.",
	},
}

func main() {
	out := flag.String("o", "aquarium.gif", "куда сохранить GIF")
	seed := flag.Uint64("seed", 1, "seed симуляции")
	frames := flag.Int("frames", 150, "сколько кадров (один шаг на кадр)")
	skip := flag.Int("skip", 300, "сколько шагов пропустить до первого кадра")
	flag.Parse()

	if err := run(*out, *seed, *frames, *skip); err != nil {
		fmt.Fprintln(os.Stderr, "gif:", err)
		os.Exit(1)
	}
}

func run(out string, seed uint64, frames, skip int) error {
	cfg := sim.DefaultConfig()
	cfg.Seed = seed
	w := sim.New(cfg)
	for range skip {
		w.Step()
	}

	anim := &gif.GIF{}
	var prev *image.Paletted
	for range frames {
		frame := draw(w)
		anim.Image = append(anim.Image, diff(prev, frame))
		anim.Delay = append(anim.Delay, 10) // в сотых долях секунды
		anim.Disposal = append(anim.Disposal, gif.DisposalNone)
		prev = frame
		w.Step()
	}

	f, err := os.Create(out)
	if err != nil {
		return err
	}
	if err := gif.EncodeAll(f, anim); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// draw рисует мир целиком.
func draw(w *sim.World) *image.Paletted {
	width := w.Width*cellW + 4*margin
	height := w.Height*cellH + 4*margin
	img := image.NewPaletted(image.Rect(0, 0, width, height), palette)
	fill(img, img.Bounds(), background)
	frame := image.Rect(margin, margin, width-margin, height-margin)
	fill(img, frame, border)
	fill(img, frame.Inset(1), background)

	// Порядок как в терминале: кто нарисован позже, тот и виден.
	for p := range w.Food {
		glyph(img, p, food)
	}
	for _, f := range w.Fish {
		glyph(img, f.Pos, fish)
	}
	for _, p := range w.Predators {
		glyph(img, p.Pos, predator)
	}
	return img
}

func glyph(img *image.Paletted, p sim.Pos, kind uint8) {
	x0 := 2*margin + p.X*cellW + (cellW-5*scale)/2
	y0 := 2*margin + p.Y*cellH + (cellH-7*scale)/2
	for y, row := range glyphs[kind] {
		for x, c := range row {
			if c == '#' {
				fill(img, image.Rect(x0+x*scale, y0+y*scale, x0+(x+1)*scale, y0+(y+1)*scale), kind)
			}
		}
	}
}

func fill(img *image.Paletted, r image.Rectangle, c uint8) {
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			img.SetColorIndex(x, y, c)
		}
	}
}

// diff оставляет в кадре только изменившиеся пиксели, остальные делает прозрачными:
// при DisposalNone сквозь них виден предыдущий кадр.
func diff(prev, cur *image.Paletted) *image.Paletted {
	if prev == nil {
		return cur
	}
	out := image.NewPaletted(cur.Rect, palette)
	for i, c := range cur.Pix {
		if c != prev.Pix[i] {
			out.Pix[i] = c
		}
	}
	return out
}
