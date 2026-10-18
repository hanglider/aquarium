//go:build js && wasm

// Command web — та же симуляция, собранная в WebAssembly для сайта.
// Страница вызывает функции из объекта aquarium и сама рисует мир на canvas.
package main

import (
	"syscall/js"

	"github.com/hanglider/aquarium/internal/sim"
)

var world *sim.World

func main() {
	js.Global().Set("aquarium", js.ValueOf(map[string]any{
		"reset": js.FuncOf(reset),
		"step":  js.FuncOf(step),
		"state": js.FuncOf(state),
	}))
	select {} // программа должна жить, пока открыта страница
}

// reset(seed) создаёт новый мир с настройками по умолчанию —
// такой же, как aquarium -seed seed в терминале.
func reset(_ js.Value, args []js.Value) any {
	cfg := sim.DefaultConfig()
	cfg.Seed = uint64(args[0].Float())
	world = sim.New(cfg)
	return nil
}

// step(n, every) делает до n шагов (меньше, если рыбы кончились) и возвращает
// точки для графиков с каждого every-го шага: шаг, рыб, хищников, скорость, зрение.
func step(_ js.Value, args []js.Value) any {
	n, every := args[0].Int(), args[1].Int()
	var points []any
	for range n {
		if len(world.Fish) == 0 {
			break
		}
		world.Step()
		if world.Steps%every == 0 {
			a := world.Averages()
			points = append(points, world.Steps, len(world.Fish), len(world.Predators), a.Speed, a.Vision)
		}
	}
	return points
}

// state() возвращает мир объектом; координаты — плоскими списками x, y, x, y...
func state(js.Value, []js.Value) any {
	a := world.Averages()
	return map[string]any{
		"width":     world.Width,
		"height":    world.Height,
		"steps":     world.Steps,
		"fish":      flat(len(world.Fish), func(i int) sim.Pos { return world.Fish[i].Pos }),
		"predators": flat(len(world.Predators), func(i int) sim.Pos { return world.Predators[i].Pos }),
		"food":      foodPositions(),
		"speed":     a.Speed,
		"vision":    a.Vision,
		"cost":      a.Cost,
	}
}

func flat(n int, pos func(int) sim.Pos) []any {
	out := make([]any, 0, 2*n)
	for i := range n {
		p := pos(i)
		out = append(out, p.X, p.Y)
	}
	return out
}

func foodPositions() []any {
	out := make([]any, 0, 2*len(world.Food))
	for p := range world.Food {
		out = append(out, p.X, p.Y)
	}
	return out
}
