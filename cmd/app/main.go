package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vshevchenk0/kv-storage/internal/app"
	"github.com/vshevchenk0/kv-storage/internal/config"
)

func main() {
	config, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	a, err := app.NewApp(config)
	if err != nil {
		panic(fmt.Sprintf("failed to create app: %v", err))
	}

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGINT)

	go func() {
		if err := a.Run(); err != nil {
			panic(fmt.Sprintf("error during app run: %v", err))
		}
	}()

	<-shutdownChan
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*1)
	defer cancel()
	doneChan := make(chan struct{}, 1)
	go func() {
		if err := a.Stop(doneChan); err != nil {
			panic(fmt.Sprintf("failed to gracefully shutdown: %v", err))
		}
	}()
	select {
	case <-ctx.Done():
		fmt.Print("timeout reached, failed to gracefully shutdown")
	case <-doneChan:
		fmt.Print("graceful shutdown complete")
	}
}
