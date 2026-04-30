// Package main is the entry point for the gator CLI.
package main

import (
	"fmt"
	"os"

	"github.com/Dzar87/gator/internal/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Read()
	if err != nil {
		return err
	}
	if err := cfg.SetUser("test"); err != nil {
		return err
	}
	cfg, err = config.Read()
	if err != nil {
		return err
	}
	payload, err := cfg.ToPrettyJson()
	if err != nil {
		return err
	}
	fmt.Println(payload)
	return nil
}
