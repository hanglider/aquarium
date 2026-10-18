package sim

import "math/rand/v2"

// Config задаёт параметры мира.
type Config struct {
	Width, Height int
	Fish          int     // начальное число рыб
	Food          int     // начальное число корма
	FoodChance    float64 // вероятность появления корма за шаг
	StartEnergy   int     // энергия новой рыбы
	FoodEnergy    int     // энергия за съеденный корм
	MoveCost      int     // расход энергии за шаг
	Seed          uint64
}

// DefaultConfig возвращает конфиг с разумными значениями.
func DefaultConfig() Config {
	return Config{
		Width:       60,
		Height:      20,
		Fish:        20,
		Food:        200,
		FoodChance:  0.5,
		StartEnergy: 60,
		FoodEnergy:  20,
		MoveCost:    1,
	}
}

// Pos — клетка сетки.
type Pos struct{ X, Y int }

// Fish — одна рыба.
type Fish struct {
	Pos    Pos
	Energy int
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
		w.Fish = append(w.Fish, Fish{Pos: w.randomPos(), Energy: cfg.StartEnergy})
	}
	for range cfg.Food {
		w.Food[w.randomPos()] = true
	}
	return w
}

func (w *World) randomPos() Pos {
	return Pos{X: w.rng.IntN(w.Width), Y: w.rng.IntN(w.Height)}
}
