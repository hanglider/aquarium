package sim

import (
	"math"
	"reflect"
	"testing"
)

// tiny — мир 1×1: рыбе некуда плыть, поэтому исход шага предсказуем.
func tiny(energy float64) *World {
	cfg := DefaultConfig()
	cfg.Width, cfg.Height = 1, 1
	cfg.Fish, cfg.Food = 1, 0
	cfg.FoodRate = 0
	cfg.StartEnergy = energy
	return New(cfg)
}

// line — мир 10×1 с одной рыбой в левой клетке и без случайного корма.
func line(g Genes) *World {
	cfg := DefaultConfig()
	cfg.Width, cfg.Height = 10, 1
	cfg.Fish, cfg.Food = 0, 0
	cfg.FoodRate = 0
	w := New(cfg)
	w.Fish = []Fish{{Pos: Pos{0, 0}, Energy: 50, Genes: g}}
	return w
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestStepEat(t *testing.T) {
	w := tiny(10)
	w.Food[Pos{0, 0}] = true
	w.Step()

	if len(w.Food) != 0 {
		t.Error("корм не съеден")
	}
	want := 10 + w.cfg.FoodEnergy - w.Cost(w.cfg.StartGenes)
	if got := w.Fish[0].Energy; !near(got, want) {
		t.Errorf("энергия %v, want %v", got, want)
	}
}

func TestStepDie(t *testing.T) {
	w := tiny(0)
	w.Fish[0].Energy = w.Cost(w.cfg.StartGenes)
	w.Step()

	if len(w.Fish) != 0 {
		t.Errorf("рыба с нулевой энергией жива: %+v", w.Fish)
	}
}

func TestStepBounds(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Width, cfg.Height = 3, 2
	cfg.BaseCost, cfg.SpeedCost, cfg.VisionCost = 0, 0, 0
	cfg.StartGenes = Genes{Speed: 3, Vision: 1}
	w := New(cfg)

	for range 1000 {
		w.Step()
		for _, f := range w.Fish {
			if !w.inside(f.Pos) {
				t.Fatalf("шаг %d: рыба вне мира: %+v", w.Steps, f.Pos)
			}
		}
	}
}

func TestStepFoodSpawn(t *testing.T) {
	w := tiny(10)
	w.cfg.FoodRate = 1
	w.Fish = nil
	w.Step()

	if len(w.Food) != 1 {
		t.Errorf("корма %d, want 1", len(w.Food))
	}
}

func TestSeeFood(t *testing.T) {
	w := line(Genes{Speed: 1, Vision: 5})
	w.Food[Pos{3, 0}] = true

	w.Step()
	if got := w.Fish[0].Pos; got != (Pos{1, 0}) {
		t.Fatalf("после шага рыба в %+v, want {1 0}: должна плыть к корму", got)
	}
	w.Step()
	w.Step()
	if len(w.Food) != 0 {
		t.Error("за 3 шага рыба не доплыла до корма")
	}
}

func TestFoodOutOfSight(t *testing.T) {
	w := line(Genes{Speed: 1, Vision: 2})
	w.Food[Pos{9, 0}] = true

	for range 3 {
		w.Step()
		if w.Fish[0].Pos.X > 3 {
			t.Fatalf("рыба уплыла в %+v, хотя корм дальше зрения", w.Fish[0].Pos)
		}
	}
}

func TestSpeed(t *testing.T) {
	w := line(Genes{Speed: 3, Vision: 5})
	w.Food[Pos{3, 0}] = true
	w.Step()

	if len(w.Food) != 0 {
		t.Errorf("рыба со скоростью 3 не доплыла за шаг, она в %+v", w.Fish[0].Pos)
	}
}

func TestCost(t *testing.T) {
	cheap := Genes{Speed: 1, Vision: 1}
	w := New(DefaultConfig())
	if w.Cost(Genes{Speed: 2, Vision: 1}) <= w.Cost(cheap) {
		t.Error("скорость должна стоить энергии")
	}
	if w.Cost(Genes{Speed: 1, Vision: 5}) <= w.Cost(cheap) {
		t.Error("зрение должно стоить энергии")
	}
}

func TestStepSplit(t *testing.T) {
	w := tiny(0)
	cost := w.Cost(w.cfg.StartGenes)
	w.Fish[0].Energy = w.cfg.SplitEnergy + cost + 1
	w.Step()

	if len(w.Fish) != 2 {
		t.Fatalf("рыб %d, want 2", len(w.Fish))
	}
	a, b := w.Fish[0].Energy, w.Fish[1].Energy
	if want := w.cfg.SplitEnergy + 1; !near(a+b, want) {
		t.Errorf("энергия %v+%v=%v, want %v (энергия не должна появляться из ничего)", a, b, a+b, want)
	}
	if !near(a, b) {
		t.Errorf("энергия поделена неровно: %v и %v", a, b)
	}
}

func TestStepNoSplitBelowThreshold(t *testing.T) {
	w := tiny(0)
	w.Fish[0].Energy = w.cfg.SplitEnergy + w.Cost(w.cfg.StartGenes) - 1
	w.Step()

	if len(w.Fish) != 1 {
		t.Errorf("рыб %d, want 1", len(w.Fish))
	}
}

func TestMutation(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Seed = 1
	w := New(cfg)
	parent := Genes{Speed: 2, Vision: 4}

	same := 0
	for range 100 {
		g := w.mutate(parent)
		if g == parent {
			same++
		}
		if math.Abs(g.Speed-parent.Speed) > 1 || math.Abs(g.Vision-parent.Vision) > 2 {
			t.Fatalf("слишком сильная мутация: %+v → %+v", parent, g)
		}
	}
	if same == 100 {
		t.Error("гены никогда не меняются")
	}

	w.cfg.Mutation = 0
	if g := w.mutate(parent); g != parent {
		t.Errorf("при Mutation=0 гены изменились: %+v", g)
	}
}

func TestDeterministic(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Seed = 42
	a, b := New(cfg), New(cfg)
	for range 200 {
		a.Step()
		b.Step()
	}

	if !reflect.DeepEqual(a.Fish, b.Fish) || !reflect.DeepEqual(a.Food, b.Food) {
		t.Error("один seed — разные миры")
	}
}
