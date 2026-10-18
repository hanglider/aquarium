package sim

import "testing"

func TestNew(t *testing.T) {
	cfg := DefaultConfig()
	w := New(cfg)

	if len(w.Fish) != cfg.Fish {
		t.Errorf("рыб %d, want %d", len(w.Fish), cfg.Fish)
	}
	if len(w.Food) == 0 || len(w.Food) > cfg.Food {
		t.Errorf("корма %d, want 1..%d", len(w.Food), cfg.Food)
	}
	for _, f := range w.Fish {
		if !w.inside(f.Pos) {
			t.Errorf("рыба вне мира: %+v", f.Pos)
		}
		if f.Energy != cfg.StartEnergy {
			t.Errorf("энергия %d, want %d", f.Energy, cfg.StartEnergy)
		}
	}
}

func (w *World) inside(p Pos) bool {
	return p.X >= 0 && p.X < w.Width && p.Y >= 0 && p.Y < w.Height
}
