package sim

// Step продвигает мир на один шаг: рыбы плывут, едят, тратят энергию,
// голодные умирают, сытые делятся, появляется новый корм.
func (w *World) Step() {
	alive := w.Fish[:0]
	var born []Fish
	for _, f := range w.Fish {
		for range w.moves(f.Genes.Speed) {
			f.Pos = w.move(f.Pos, f.Genes.Vision)
			if w.Food[f.Pos] {
				delete(w.Food, f.Pos)
				f.Energy += w.cfg.FoodEnergy
			}
		}
		f.Energy -= w.Cost(f.Genes)
		if f.Energy <= 0 {
			continue
		}
		if f.Energy >= w.cfg.SplitEnergy {
			child := Fish{Pos: f.Pos, Energy: f.Energy / 2, Genes: w.mutate(f.Genes)}
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

// moves — сколько клеток рыба проплывёт за шаг: целая часть скорости всегда,
// дробная — с такой вероятностью.
func (w *World) moves(speed float64) int {
	n := int(speed)
	if w.rng.Float64() < speed-float64(n) {
		n++
	}
	return n
}

// spawnFood добавляет в среднем FoodRate корма.
func (w *World) spawnFood() {
	for range w.moves(w.cfg.FoodRate) {
		w.Food[w.randomPos()] = true
	}
}

// move сдвигает рыбу на одну клетку: к ближайшему видимому корму,
// а если корма не видно — в случайную сторону (или оставляет на месте).
func (w *World) move(p Pos, vision float64) Pos {
	var dx, dy int
	if food, ok := w.nearestFood(p, vision); ok {
		dx, dy = sign(food.X-p.X), sign(food.Y-p.Y)
	} else {
		dx, dy = w.rng.IntN(3)-1, w.rng.IntN(3)-1
	}
	p.X = clamp(p.X+dx, 0, w.Width-1)
	p.Y = clamp(p.Y+dy, 0, w.Height-1)
	return p
}

// nearestFood ищет ближайший корм в квадрате с радиусом vision.
// Расстояние — число ходов по 8 направлениям, то есть max(|dx|, |dy|).
// Клетки перебираются в фиксированном порядке, поэтому при равных
// расстояниях выбор всегда один и тот же — симуляция остаётся детерминированной.
func (w *World) nearestFood(p Pos, vision float64) (Pos, bool) {
	r := int(vision)
	best, bestDist := Pos{}, r+1
	for y := max(0, p.Y-r); y <= min(w.Height-1, p.Y+r); y++ {
		for x := max(0, p.X-r); x <= min(w.Width-1, p.X+r); x++ {
			q := Pos{X: x, Y: y}
			if d := max(abs(x-p.X), abs(y-p.Y)); d < bestDist && w.Food[q] {
				best, bestDist = q, d
			}
		}
	}
	return best, bestDist <= r
}

func clamp(v, lo, hi int) int {
	return max(lo, min(v, hi))
}

func sign(v int) int {
	switch {
	case v > 0:
		return 1
	case v < 0:
		return -1
	}
	return 0
}

func abs(v int) int {
	return max(v, -v)
}
