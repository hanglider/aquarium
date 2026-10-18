// Command aquarium — консольный симулятор эволюции рыб.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/hanglider/aquarium/internal/render"
	"github.com/hanglider/aquarium/internal/sim"
	"github.com/hanglider/aquarium/internal/stats"
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
	steps := flag.Int("steps", 0, "прогнать столько шагов без экрана и выйти")
	csvPath := flag.String("csv", "", "писать статистику каждого шага в этот CSV-файл")
	flag.Uint64Var(&cfg.Seed, "seed", 0, "seed симуляции, 0 — случайный")
	showVersion := flag.Bool("version", false, "показать версию и выйти")
	flag.Parse()

	if *showVersion {
		fmt.Println("aquarium", version)
		return
	}

	if cfg.Width < 1 || cfg.Height < 1 || cfg.Fish < 0 || cfg.Predators < 0 || *speed < 1 || *steps < 0 {
		fmt.Fprintln(os.Stderr, "width, height и speed должны быть больше 0, fish, predators и steps — не меньше 0")
		os.Exit(2)
	}
	random := cfg.Seed == 0
	if random {
		cfg.Seed = rand.Uint64()
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := start(ctx, cfg, *speed, *steps, *csvPath); err != nil {
		fmt.Fprintln(os.Stderr, "aquarium:", err)
		os.Exit(1)
	}
	fmt.Printf("seed %d, повторить: %s\n", cfg.Seed, repeat(os.Args[1:], cfg.Seed, random))
}

// repeat собирает команду, которая повторит этот запуск. Если seed был
// случайным, он дописывается в конец: из нескольких -seed побеждает последний.
func repeat(args []string, seed uint64, random bool) string {
	cmd := append([]string{"aquarium"}, args...)
	if random {
		cmd = append(cmd, "-seed", strconv.FormatUint(seed, 10))
	}
	return strings.Join(cmd, " ")
}

// start открывает CSV, если он нужен, и запускает симуляцию: на экране
// или, если задано steps, без экрана.
func start(ctx context.Context, cfg sim.Config, speed, steps int, csvPath string) (err error) {
	out := io.Discard // без -csv строки пишутся в никуда — так коду не нужны проверки
	if csvPath != "" {
		f, ferr := os.Create(csvPath) // не err: иначе defer ниже писал бы в затенённую переменную
		if ferr != nil {
			return ferr
		}
		defer func() { err = errors.Join(err, f.Close()) }()
		out = f
	}
	rec, err := stats.NewCSV(out)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, rec.Close()) }()

	w := sim.New(cfg)
	if err := rec.Row(w); err != nil {
		return err
	}
	if steps > 0 {
		return headless(ctx, w, steps, rec)
	}
	return run(ctx, w, speed, rec)
}

// headless прогоняет steps шагов без экрана и печатает итог.
func headless(ctx context.Context, w *sim.World, steps int, rec *stats.CSV) error {
	for range steps {
		if ctx.Err() != nil || len(w.Fish) == 0 {
			break
		}
		w.Step()
		if err := rec.Row(w); err != nil {
			return err
		}
	}
	fmt.Printf("шаг %d  рыб %d  хищников %d  корма %d\n", w.Steps, len(w.Fish), len(w.Predators), len(w.Food))
	return nil
}

// run показывает аквариум на экране, пока его не закроют.
func run(ctx context.Context, w *sim.World, speed int, rec *stats.CSV) error {
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

	step := func() error {
		w.Step()
		return rec.Row(w)
	}
	paused := false
	for {
		status, steps := "пробел — пауза  q — выход", tick.C
		// Из nil-канала ничего не придёт: пока он стоит в select, симуляция не идёт.
		switch {
		case len(w.Fish) == 0:
			status, steps = fmt.Sprintf("Все рыбы умерли на шаге %d. Выход — q.", w.Steps), nil
		case paused:
			status, steps = "пауза: пробел — продолжить  n — шаг  q — выход", nil
		}
		render.Frame(scr, w, status)

		select {
		case <-ctx.Done():
			return nil
		case ev := <-events:
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch {
				case ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC || ev.Rune() == 'q':
					return nil
				case ev.Rune() == ' ':
					paused = !paused
				case ev.Rune() == 'n' && paused && len(w.Fish) > 0:
					if err := step(); err != nil {
						return err
					}
				}
			case *tcell.EventResize:
				scr.Sync()
			}
		case <-steps:
			if err := step(); err != nil {
				return err
			}
		}
	}
}
