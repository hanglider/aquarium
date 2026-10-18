package render

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/hanglider/aquarium/internal/sim"
)

var ansi = regexp.MustCompile(`\x1b\[[0-9;?]*[A-Za-z]`)

func TestFrame(t *testing.T) {
	cfg := sim.DefaultConfig()
	cfg.Width, cfg.Height = 4, 2
	cfg.Fish, cfg.Food = 0, 0
	w := sim.New(cfg)
	w.Fish = []sim.Fish{{Pos: sim.Pos{X: 0, Y: 0}, Energy: 1, Genes: sim.Genes{Speed: 1.5, Vision: 3}}}
	w.Food[sim.Pos{X: 3, Y: 1}] = true

	var buf bytes.Buffer
	Frame(&buf, w, 42)

	want := "" +
		"┌────┐\n" +
		"│>   │\n" +
		"│   *│\n" +
		"└────┘\n" +
		"шаг 0  рыб 1  корма 1  seed 42\n" +
		"в среднем: скорость 1.50  зрение 3.0  расход 1.60\n"
	if got := ansi.ReplaceAllString(buf.String(), ""); got != want {
		t.Errorf("кадр:\n%s\nwant:\n%s", got, want)
	}
}
