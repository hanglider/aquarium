// Package render отрисовывает состояние симуляции в терминал.
package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/hanglider/aquarium/internal/sim"
)

// ANSI escape-коды.
const (
	home       = "\x1b[H"
	clear      = "\x1b[2J"
	clearLine  = "\x1b[K"
	hideCursor = "\x1b[?25l"
	showCursor = "\x1b[?25h"
	reset      = "\x1b[0m"
	yellow     = "\x1b[33m"
	green      = "\x1b[32m"
	cyan       = "\x1b[36m"
)

const (
	fishCell = yellow + ">" + reset
	foodCell = green + "*" + reset
)

// Start готовит терминал: очищает экран и прячет курсор.
func Start(out io.Writer) {
	fmt.Fprint(out, clear, hideCursor)
}

// Stop возвращает терминал в обычное состояние.
func Stop(out io.Writer) {
	fmt.Fprint(out, reset, showCursor)
}

// Frame рисует один кадр: аквариум в рамке и строку статистики.
// Курсор переводится в начало экрана, а не экран очищается, — так нет мерцания.
func Frame(out io.Writer, w *sim.World, seed uint64) {
	fish := make(map[sim.Pos]bool, len(w.Fish))
	for _, f := range w.Fish {
		fish[f.Pos] = true
	}

	var b strings.Builder
	b.WriteString(home)
	border := cyan + strings.Repeat("─", w.Width)
	b.WriteString(cyan + "┌" + border + "┐" + reset + "\n")
	for y := range w.Height {
		b.WriteString(cyan + "│" + reset)
		for x := range w.Width {
			p := sim.Pos{X: x, Y: y}
			switch {
			case fish[p]:
				b.WriteString(fishCell)
			case w.Food[p]:
				b.WriteString(foodCell)
			default:
				b.WriteByte(' ')
			}
		}
		b.WriteString(cyan + "│" + reset + "\n")
	}
	b.WriteString(cyan + "└" + border + "┘" + reset + "\n")
	fmt.Fprintf(&b, "шаг %d  рыб %d  корма %d  seed %d%s\n",
		w.Steps, len(w.Fish), len(w.Food), seed, clearLine)
	a := w.Averages()
	fmt.Fprintf(&b, "в среднем: скорость %.2f  зрение %.1f  расход %.2f%s\n",
		a.Speed, a.Vision, a.Cost, clearLine)

	io.WriteString(out, b.String())
}
