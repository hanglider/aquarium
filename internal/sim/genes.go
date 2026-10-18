package sim

// Genes — наследуемые свойства рыбы.
type Genes struct {
	Speed  float64 // клеток за шаг; дробная часть — шанс проплыть ещё одну
	Vision float64 // радиус, в котором рыба видит корм
}

// Границы генов: без них мутации за тысячи шагов уводят значения в бессмыслицу.
const (
	minSpeed, maxSpeed   = 0.1, 5
	minVision, maxVision = 0.5, 15
)

// Cost — расход энергии за шаг для рыбы с такими генами.
// Скорость дорожает квадратично, зрение — линейно: за всё приходится платить.
func (w *World) Cost(g Genes) float64 {
	c := w.cfg
	return c.BaseCost + c.SpeedCost*g.Speed*g.Speed + c.VisionCost*g.Vision
}

// mutate возвращает гены малька: каждый ген родителя умножается
// на случайный множитель около 1 (разброс задаёт Config.Mutation).
func (w *World) mutate(g Genes) Genes {
	g.Speed = clampF(g.Speed*(1+w.rng.NormFloat64()*w.cfg.Mutation), minSpeed, maxSpeed)
	g.Vision = clampF(g.Vision*(1+w.rng.NormFloat64()*w.cfg.Mutation), minVision, maxVision)
	return g
}

func clampF(v, lo, hi float64) float64 {
	return max(lo, min(v, hi))
}

// Averages — средние значения генов и расхода энергии по живым рыбам.
type Averages struct {
	Speed, Vision, Cost float64
}

// Averages считает средние по популяции; для пустого аквариума — нули.
func (w *World) Averages() Averages {
	var a Averages
	if len(w.Fish) == 0 {
		return a
	}
	for _, f := range w.Fish {
		a.Speed += f.Genes.Speed
		a.Vision += f.Genes.Vision
		a.Cost += w.Cost(f.Genes)
	}
	n := float64(len(w.Fish))
	a.Speed /= n
	a.Vision /= n
	a.Cost /= n
	return a
}
