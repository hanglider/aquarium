package render

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/hanglider/aquarium/internal/sim"
)

// screen — виртуальный терминал tcell: в него можно рисовать без настоящего экрана.
func screen(t *testing.T, width, height int) tcell.SimulationScreen {
	t.Helper()
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Fini)
	s.SetSize(width, height)
	return s
}

// text возвращает содержимое экрана построчно, без пробелов в конце строк.
func text(s tcell.Screen) string {
	width, height := s.Size()
	var b strings.Builder
	for y := range height {
		var row strings.Builder
		for x := range width {
			str, _, _ := s.Get(x, y)
			row.WriteString(str)
		}
		b.WriteString(strings.TrimRight(row.String(), " ") + "\n")
	}
	return b.String()
}

func TestFrame(t *testing.T) {
	cfg := sim.DefaultConfig()
	cfg.Width, cfg.Height = 4, 2
	cfg.Fish, cfg.Food, cfg.Predators = 0, 0, 0
	w := sim.New(cfg)
	w.Fish = []sim.Fish{{Pos: sim.Pos{X: 0, Y: 0}, Energy: 1, Genes: sim.Genes{Speed: 1.5, Vision: 3}}}
	w.Predators = []sim.Fish{{Pos: sim.Pos{X: 1, Y: 1}, Energy: 1}}
	w.Food[sim.Pos{X: 3, Y: 1}] = true

	s := screen(t, 60, 7)
	Frame(s, w, 42, "конец")

	want := "" +
		"┌────┐\n" +
		"│>   │\n" +
		"│ @ *│\n" +
		"└────┘\n" +
		"шаг 0  рыб 1  хищников 1  корма 1  seed 42\n" +
		"в среднем: скорость 1.50  зрение 3.0  расход 1.60\n" +
		"конец\n"
	if got := text(s); got != want {
		t.Errorf("кадр:\n%s\nwant:\n%s", got, want)
	}

	if _, style, _ := s.Get(2, 2); style != predatorStyle {
		t.Errorf("хищник нарисован не тем цветом: %v", style)
	}
}
