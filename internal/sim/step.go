package sim

// Step продвигает мир на один шаг: рыбы плывут, едят, тратят энергию,
// голодные умирают, появляется новый корм.
func (w *World) Step() {
	alive := w.Fish[:0]
	for _, f := range w.Fish {
		f.Pos = w.move(f.Pos)
		if w.Food[f.Pos] {
			delete(w.Food, f.Pos)
			f.Energy += w.cfg.FoodEnergy
		}
		f.Energy -= w.cfg.MoveCost
		if f.Energy > 0 {
			alive = append(alive, f)
		}
	}
	w.Fish = alive

	if w.rng.Float64() < w.cfg.FoodChance {
		w.Food[w.randomPos()] = true
	}
	w.Steps++
}

// move сдвигает на случайную соседнюю клетку (или оставляет на месте),
// не выпуская за край.
func (w *World) move(p Pos) Pos {
	p.X = clamp(p.X+w.rng.IntN(3)-1, 0, w.Width-1)
	p.Y = clamp(p.Y+w.rng.IntN(3)-1, 0, w.Height-1)
	return p
}

func clamp(v, lo, hi int) int {
	return max(lo, min(v, hi))
}
