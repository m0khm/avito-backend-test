package main

import (
	"fmt"
	"log"

	"room-booking-service/internal/app"
	"room-booking-service/internal/config"
)

type runner interface {
	Run() error
}

func runMain(load func() config.Config, build func(config.Config) (runner, error)) error {
	cfg := load()
	application, err := build(cfg)
	if err != nil {
		return fmt.Errorf("app init failed: %w", err)
	}
	if err := application.Run(); err != nil {
		return fmt.Errorf("app run failed: %w", err)
	}
	return nil
}

func main() {
	if err := runMain(config.Load, func(cfg config.Config) (runner, error) {
		return app.New(cfg)
	}); err != nil {
		log.Fatal(err)
	}
}
