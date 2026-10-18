// Package stats записывает статистику симуляции в CSV.
package stats

import (
	"encoding/csv"
	"io"
	"strconv"

	"github.com/hanglider/aquarium/internal/sim"
)

var header = []string{"step", "fish", "predators", "food", "speed", "vision", "cost"}

// CSV пишет по строке на каждый шаг симуляции.
type CSV struct {
	w *csv.Writer
}

// NewCSV создаёт запись в out и сразу пишет строку с названиями колонок.
func NewCSV(out io.Writer) (*CSV, error) {
	c := &CSV{w: csv.NewWriter(out)}
	return c, c.w.Write(header)
}

// Row добавляет строку с текущим состоянием мира.
func (c *CSV) Row(w *sim.World) error {
	a := w.Averages()
	return c.w.Write([]string{
		strconv.Itoa(w.Steps),
		strconv.Itoa(len(w.Fish)),
		strconv.Itoa(len(w.Predators)),
		strconv.Itoa(len(w.Food)),
		num(a.Speed),
		num(a.Vision),
		num(a.Cost),
	})
}

// Close дописывает буфер; без него последние строки не попадут в файл.
func (c *CSV) Close() error {
	c.w.Flush()
	return c.w.Error()
}

func num(v float64) string {
	return strconv.FormatFloat(v, 'f', 3, 64)
}
