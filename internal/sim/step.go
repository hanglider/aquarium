package sim

// Step продвигает мир на один шаг: рыбы плывут и едят корм, хищники плывут
// и едят рыб, голодные умирают, сытые делятся, появляется новый корм.
func (w *World) Step() {
	w.stepFish()
	w.stepPredators()
	w.spawnFood()
	w.spawnPredators()
	w.Steps++
}

// stepFish — ход рыб. Рыба, которая видит хищника, уплывает от него,
// иначе плывёт к корму.
func (w *World) stepFish() {
	danger := make(map[Pos]bool, len(w.Predators))
	for _, p := range w.Predators {
		danger[p.Pos] = true
	}
	alive := w.Fish[:0]
	var born []Fish
	for _, f := range w.Fish {
		for range w.moves(f.Genes.Speed) {
			f.Pos = w.swim(f.Pos, f.Genes.Vision, danger)
			if w.Food[f.Pos] {
				delete(w.Food, f.Pos)
				f.Energy += w.cfg.FoodEnergy
			}
		}
		w.live(f, w.cfg.SplitEnergy, &alive, &born)
	}
	// Мальки добавляются после цикла: alive пишет в тот же массив,
	// что и w.Fish, и append внутри цикла затёр бы ещё не обработанных рыб.
	w.Fish = append(alive, born...)
}

// stepPredators — ход хищников: плывут к ближайшей видимой рыбе
// и съедают по одной рыбе в каждой клетке, куда попали.
func (w *World) stepPredators() {
	prey := make(map[Pos][]int, len(w.Fish)) // клетка → номера рыб в ней
	for i, f := range w.Fish {
		prey[f.Pos] = append(prey[f.Pos], i)
	}
	hasPrey := func(p Pos) bool { return len(prey[p]) > 0 }
	eaten := make([]bool, len(w.Fish))

	alive := w.Predators[:0]
	var born []Fish
	for _, p := range w.Predators {
		for range w.moves(p.Genes.Speed) {
			p.Pos = w.toward(p.Pos, p.Genes.Vision, hasPrey)
			if fish := prey[p.Pos]; len(fish) > 0 {
				eaten[fish[len(fish)-1]] = true
				prey[p.Pos] = fish[:len(fish)-1]
				p.Energy += w.cfg.PreyEnergy
			}
		}
		w.live(p, w.cfg.PredatorSplit, &alive, &born)
	}
	w.Predators = append(alive, born...)

	survivors := w.Fish[:0]
	for i, f := range w.Fish {
		if !eaten[i] {
			survivors = append(survivors, f)
		}
	}
	w.Fish = survivors
}

// live списывает с особи расход энергии за шаг. Умершая никуда не попадает,
// живая — в alive, а если накопила split энергии, делится: малёк с мутацией — в born.
func (w *World) live(f Fish, split float64, alive, born *[]Fish) {
	f.Energy -= w.Cost(f.Genes)
	if f.Energy <= 0 {
		return
	}
	if f.Energy >= split {
		child := Fish{Pos: f.Pos, Energy: f.Energy / 2, Genes: w.mutate(f.Genes)}
		f.Energy -= child.Energy
		*born = append(*born, child)
	}
	*alive = append(*alive, f)
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

// spawnPredators — в среднем PredatorRate новых хищников за шаг приплывает извне.
// Без этого хищники рано или поздно вымирают и больше не появляются.
func (w *World) spawnPredators() {
	for range w.moves(w.cfg.PredatorRate) {
		w.Predators = append(w.Predators, Fish{Pos: w.randomPos(), Energy: w.cfg.PredatorEnergy, Genes: w.cfg.PredatorGenes})
	}
}

// swim сдвигает рыбу на одну клетку: от ближайшего видимого хищника,
// если хищника не видно — к ближайшему видимому корму, иначе — куда попало.
func (w *World) swim(p Pos, vision float64, danger map[Pos]bool) Pos {
	if enemy, ok := w.nearest(p, vision, func(q Pos) bool { return danger[q] }); ok && enemy != p {
		return w.shift(p, sign(p.X-enemy.X), sign(p.Y-enemy.Y))
	}
	return w.toward(p, vision, func(q Pos) bool { return w.Food[q] })
}

// toward сдвигает на одну клетку к ближайшей видимой клетке, где есть цель,
// а если такой не видно — в случайную сторону (или оставляет на месте).
func (w *World) toward(p Pos, vision float64, target func(Pos) bool) Pos {
	if q, ok := w.nearest(p, vision, target); ok {
		return w.shift(p, sign(q.X-p.X), sign(q.Y-p.Y))
	}
	return w.shift(p, w.rng.IntN(3)-1, w.rng.IntN(3)-1)
}

// shift сдвигает позицию на (dx, dy), не выпуская за край аквариума.
func (w *World) shift(p Pos, dx, dy int) Pos {
	p.X = clamp(p.X+dx, 0, w.Width-1)
	p.Y = clamp(p.Y+dy, 0, w.Height-1)
	return p
}

// nearest ищет ближайшую клетку с целью в квадрате с радиусом vision.
// Расстояние — число ходов по 8 направлениям, то есть max(|dx|, |dy|).
// Клетки перебираются в фиксированном порядке, поэтому при равных
// расстояниях выбор всегда один и тот же — симуляция остаётся детерминированной.
func (w *World) nearest(p Pos, vision float64, target func(Pos) bool) (Pos, bool) {
	r := int(vision)
	best, bestDist := Pos{}, r+1
	for y := max(0, p.Y-r); y <= min(w.Height-1, p.Y+r); y++ {
		for x := max(0, p.X-r); x <= min(w.Width-1, p.X+r); x++ {
			q := Pos{X: x, Y: y}
			if d := max(abs(x-p.X), abs(y-p.Y)); d < bestDist && target(q) {
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
