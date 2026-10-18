// Package render рисует состояние симуляции в терминале через tcell.
package render

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/hanglider/aquarium/internal/sim"
)

// Цвета — те же, что были в ANSI-версии: красный, жёлтый, зелёный, бирюзовый.
var (
	borderStyle   = tcell.StyleDefault.Foreground(tcell.ColorTeal)
	predatorStyle = tcell.StyleDefault.Foreground(tcell.ColorMaroon)
	fishStyle     = tcell.StyleDefault.Foreground(tcell.ColorOlive)
	foodStyle     = tcell.StyleDefault.Foreground(tcell.ColorGreen)
)

// Frame рисует один кадр: аквариум в рамке и две строки статистики под ним.
// tcell сам сравнивает кадр с предыдущим и перерисовывает только изменённые клетки.
func Frame(s tcell.Screen, w *sim.World, seed uint64) {
	s.Clear()

	line := strings.Repeat("─", w.Width)
	s.PutStrStyled(0, 0, "┌"+line+"┐", borderStyle)
	for y := range w.Height {
		s.PutStrStyled(0, y+1, "│", borderStyle)
		s.PutStrStyled(w.Width+1, y+1, "│", borderStyle)
	}
	s.PutStrStyled(0, w.Height+1, "└"+line+"┘", borderStyle)

	// Порядок важен: кто нарисован позже, тот и виден, если клетка общая.
	for p := range w.Food {
		cell(s, p, "*", foodStyle)
	}
	for _, f := range w.Fish {
		cell(s, f.Pos, ">", fishStyle)
	}
	for _, p := range w.Predators {
		cell(s, p.Pos, "@", predatorStyle)
	}

	s.PutStr(0, w.Height+2, fmt.Sprintf("шаг %d  рыб %d  хищников %d  корма %d  seed %d",
		w.Steps, len(w.Fish), len(w.Predators), len(w.Food), seed))
	a := w.Averages()
	s.PutStr(0, w.Height+3, fmt.Sprintf("в среднем: скорость %.2f  зрение %.1f  расход %.2f",
		a.Speed, a.Vision, a.Cost))

	s.Show()
}

// Notice выводит сообщение строкой ниже статистики.
func Notice(s tcell.Screen, w *sim.World, text string) {
	s.PutStr(0, w.Height+4, text)
	s.Show()
}

// cell рисует символ в клетке аквариума (сдвиг на 1 — из-за рамки).
func cell(s tcell.Screen, p sim.Pos, ch string, style tcell.Style) {
	s.PutStrStyled(p.X+1, p.Y+1, ch, style)
}
