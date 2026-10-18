package stats

import (
	"strings"
	"testing"

	"github.com/hanglider/aquarium/internal/sim"
)

func TestCSV(t *testing.T) {
	cfg := sim.DefaultConfig()
	cfg.Fish, cfg.Food, cfg.Predators = 0, 0, 0
	cfg.Seed = 42
	w := sim.New(cfg)
	w.Fish = []sim.Fish{{Genes: sim.Genes{Speed: 1.5, Vision: 3}}}
	w.Predators = []sim.Fish{{}, {}}
	w.Food[sim.Pos{}] = true

	var b strings.Builder
	c, err := NewCSV(&b)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Row(w); err != nil {
		t.Fatal(err)
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}

	want := "seed,step,fish,predators,food,speed,vision,cost\n" +
		"42,0,1,2,1,1.500,3.000,1.600\n"
	if got := b.String(); got != want {
		t.Errorf("CSV:\n%s\nwant:\n%s", got, want)
	}
}
