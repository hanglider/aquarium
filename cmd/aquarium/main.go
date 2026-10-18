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

	"github.com/gdamore/tcell/v2"

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

	if err := run(ctx, cfg, *speed); err != nil {
		fmt.Fprintln(os.Stderr, "aquarium:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg sim.Config, speed int) error {
	scr, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	if err := scr.Init(); err != nil {
		return err
	}
	defer scr.Fini()

	// Клавиши и изменения размера окна tcell присылает событиями в канал.
	events := make(chan tcell.Event, 16)
	quit := make(chan struct{})
	go scr.ChannelEvents(events, quit)
	defer close(quit)

	tick := time.NewTicker(time.Second / time.Duration(speed))
	defer tick.Stop()
	steps := tick.C

	w := sim.New(cfg)
	for {
		render.Frame(scr, w, cfg.Seed)
		if len(w.Fish) == 0 {
			render.Notice(scr, w, fmt.Sprintf("Все рыбы умерли на шаге %d. Выход — q.", w.Steps))
			steps = nil // из nil-канала ничего не придёт: симуляция стоит, ждём выхода
		}
		select {
		case <-ctx.Done():
			return nil
		case ev := <-events:
			switch ev := ev.(type) {
			case *tcell.EventKey:
				if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC || ev.Rune() == 'q' {
					return nil
				}
			case *tcell.EventResize:
				scr.Sync()
			}
		case <-steps:
			w.Step()
		}
	}
}
