// Command aquarium — консольный симулятор эволюции рыб.
package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hanglider/aquarium/internal/render"
	"github.com/hanglider/aquarium/internal/sim"
)

// version подставляется при сборке релиза: -ldflags "-X main.version=v1.2.3".
var version = "dev"

func main() {
	cfg := sim.DefaultConfig()
	flag.IntVar(&cfg.Width, "width", cfg.Width, "ширина аквариума")
	flag.IntVar(&cfg.Height, "height", cfg.Height, "высота аквариума")
	flag.IntVar(&cfg.Fish, "fish", cfg.Fish, "число рыб")
	flag.IntVar(&cfg.Predators, "predators", cfg.Predators, "число хищников в начале")
	speed := flag.Int("speed", 10, "шагов в секунду")
	flag.Uint64Var(&cfg.Seed, "seed", 0, "seed симуляции, 0 — случайный")
	showVersion := flag.Bool("version", false, "показать версию и выйти")
	flag.Parse()

	if *showVersion {
		fmt.Println("aquarium", version)
		return
	}

	if cfg.Width < 1 || cfg.Height < 1 || cfg.Fish < 0 || cfg.Predators < 0 || *speed < 1 {
		fmt.Fprintln(os.Stderr, "width, height и speed должны быть больше 0, fish и predators — не меньше 0")
		os.Exit(2)
	}
	if cfg.Seed == 0 {
		cfg.Seed = rand.Uint64()
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	run(ctx, cfg, *speed)
}

func run(ctx context.Context, cfg sim.Config, speed int) {
	w := sim.New(cfg)
	out := os.Stdout
	render.Start(out)
	defer render.Stop(out)

	tick := time.NewTicker(time.Second / time.Duration(speed))
	defer tick.Stop()

	for {
		render.Frame(out, w, cfg.Seed)
		if len(w.Fish) == 0 {
			fmt.Fprintf(out, "Все рыбы умерли на шаге %d.\n", w.Steps)
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			w.Step()
		}
	}
}
