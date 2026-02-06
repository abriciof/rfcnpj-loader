package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/abriciof/rfcnpj-loader/internal/app"
	"github.com/abriciof/rfcnpj-loader/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		cancel()
	}()

	if err := app.Run(ctx, cfg); err != nil {
		log.Fatalf("run: %v", err)
	}
}
