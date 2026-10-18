package sim

// Step продвигает мир на один шаг: рыбы плывут, едят, тратят энергию,
// голодные умирают, сытые делятся, появляется новый корм.
func (w *World) Step() {
	alive := w.Fish[:0]
	var born []Fish
	for _, f := range w.Fish {
		f.Pos = w.move(f.Pos)
		if w.Food[f.Pos] {
			delete(w.Food, f.Pos)
			f.Energy += w.cfg.FoodEnergy
		}
		f.Energy -= w.cfg.MoveCost
		if f.Energy <= 0 {
			continue
		}
		if f.Energy >= w.cfg.SplitEnergy {
			child := Fish{Pos: f.Pos, Energy: f.Energy / 2}
			f.Energy -= child.Energy
			born = append(born, child)
		}
		alive = append(alive, f)
	}
	// Мальки добавляются после цикла: alive пишет в тот же массив,
	// что и w.Fish, и append внутри цикла затёр бы ещё не обработанных рыб.
	w.Fish = append(alive, born...)

	w.spawnFood()
	w.Steps++
}

// spawnFood добавляет в среднем FoodRate корма: целая часть — всегда,
// дробная — с такой вероятностью.
func (w *World) spawnFood() {
	n := int(w.cfg.FoodRate)
	if w.rng.Float64() < w.cfg.FoodRate-float64(n) {
		n++
	}
	for range n {
		w.Food[w.randomPos()] = true
	}
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
