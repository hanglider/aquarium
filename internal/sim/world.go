package sim

import "math/rand/v2"

// Config задаёт параметры мира.
type Config struct {
	Width, Height int
	Fish          int     // начальное число рыб
	Food          int     // начальное число корма
	FoodRate      float64 // сколько корма появляется за шаг в среднем
	StartEnergy   float64 // энергия новой рыбы
	FoodEnergy    float64 // энергия за съеденный корм
	SplitEnergy   float64 // при такой энергии рыба делится надвое
	StartGenes    Genes   // гены рыб в начале
	Mutation      float64 // насколько сильно гены малька отличаются от родителя
	Seed          uint64

	// Расход энергии за шаг: BaseCost + SpeedCost·скорость² + VisionCost·зрение.
	BaseCost, SpeedCost, VisionCost float64
}

// DefaultConfig возвращает конфиг с разумными значениями.
func DefaultConfig() Config {
	return Config{
		Width:       60,
		Height:      20,
		Fish:        20,
		Food:        200,
		FoodRate:    3,
		StartEnergy: 60,
		FoodEnergy:  20,
		SplitEnergy: 120,
		StartGenes:  Genes{Speed: 1, Vision: 2},
		Mutation:    0.1,
		BaseCost:    0.4,
		SpeedCost:   0.4,
		VisionCost:  0.1,
	}
}

// Pos — клетка сетки.
type Pos struct{ X, Y int }

// Fish — одна рыба.
type Fish struct {
	Pos    Pos
	Energy float64
	Genes  Genes
}

// World — состояние аквариума.
type World struct {
	Width, Height int
	Fish          []Fish
	Food          map[Pos]bool
	Steps         int

	cfg Config
	rng *rand.Rand
}

// New создаёт мир по конфигу. Одинаковый Seed даёт одинаковый мир.
func New(cfg Config) *World {
	w := &World{
		Width:  cfg.Width,
		Height: cfg.Height,
		Food:   make(map[Pos]bool),
		cfg:    cfg,
		rng:    rand.New(rand.NewPCG(cfg.Seed, 0)),
	}
	for range cfg.Fish {
		w.Fish = append(w.Fish, Fish{Pos: w.randomPos(), Energy: cfg.StartEnergy, Genes: cfg.StartGenes})
	}
	for range cfg.Food {
		w.Food[w.randomPos()] = true
	}
	return w
}

func (w *World) randomPos() Pos {
	return Pos{X: w.rng.IntN(w.Width), Y: w.rng.IntN(w.Height)}
}
