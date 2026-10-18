package sim

import "testing"

func TestAverages(t *testing.T) {
	w := line(Genes{Speed: 1, Vision: 2})
	w.Fish = append(w.Fish, Fish{Genes: Genes{Speed: 3, Vision: 4}})

	a := w.Averages()
	if !near(a.Speed, 2) || !near(a.Vision, 3) {
		t.Errorf("средние %+v, want скорость 2, зрение 3", a)
	}
	wantCost := (w.Cost(Genes{Speed: 1, Vision: 2}) + w.Cost(Genes{Speed: 3, Vision: 4})) / 2
	if !near(a.Cost, wantCost) {
		t.Errorf("средний расход %v, want %v", a.Cost, wantCost)
	}

	w.Fish = nil
	if a := w.Averages(); a != (Averages{}) {
		t.Errorf("без рыб средние %+v, want нули", a)
	}
}
