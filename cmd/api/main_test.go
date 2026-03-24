package main

import (
	"errors"
	"testing"

	"room-booking-service/internal/config"
)

type fakeRunner struct{ err error }

func (f fakeRunner) Run() error { return f.err }

func TestRunMain_Success(t *testing.T) {
	err := runMain(
		func() config.Config { return config.Config{HTTPPort: "8080"} },
		func(cfg config.Config) (runner, error) { return fakeRunner{}, nil },
	)
	if err != nil {
		t.Fatalf("runMain() error = %v", err)
	}
}

func TestRunMain_BuildError(t *testing.T) {
	err := runMain(
		func() config.Config { return config.Config{} },
		func(cfg config.Config) (runner, error) { return nil, errors.New("build failed") },
	)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunMain_RunError(t *testing.T) {
	err := runMain(
		func() config.Config { return config.Config{} },
		func(cfg config.Config) (runner, error) { return fakeRunner{err: errors.New("run failed")}, nil },
	)
	if err == nil {
		t.Fatal("expected error")
	}
}
