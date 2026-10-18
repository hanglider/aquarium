package sim

import (
	"reflect"
	"testing"
)

// tiny — мир 1×1: рыбе некуда плыть, поэтому исход шага предсказуем.
func tiny(energy int) *World {
	cfg := DefaultConfig()
	cfg.Width, cfg.Height = 1, 1
	cfg.Fish, cfg.Food = 1, 0
	cfg.FoodChance = 0
	cfg.StartEnergy = energy
	return New(cfg)
}

func TestStepEat(t *testing.T) {
	w := tiny(10)
	w.Food[Pos{0, 0}] = true
	w.Step()

	if len(w.Food) != 0 {
		t.Error("корм не съеден")
	}
	want := 10 + w.cfg.FoodEnergy - w.cfg.MoveCost
	if got := w.Fish[0].Energy; got != want {
		t.Errorf("энергия %d, want %d", got, want)
	}
}

func TestStepDie(t *testing.T) {
	w := tiny(1)
	w.Step()

	if len(w.Fish) != 0 {
		t.Errorf("рыба с нулевой энергией жива: %+v", w.Fish)
	}
}

func TestStepBounds(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Width, cfg.Height = 3, 2
	cfg.MoveCost = 0
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
	w.cfg.FoodChance = 1
	w.Fish = nil
	w.Step()

	if len(w.Food) != 1 {
		t.Errorf("корма %d, want 1", len(w.Food))
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
