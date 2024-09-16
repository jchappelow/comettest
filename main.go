package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jchappelow/comettest/app"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		cancel()
	}()

	if err := mainCore(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func mainCore(ctx context.Context) error {
	if len(os.Args) < 2 {
		return errors.New("specify db dir")
	}
	dbDir := os.Args[1]

	port := "26658"
	if len(os.Args) > 2 {
		port = os.Args[2]
	}

	server, err := app.StartAppServer(ctx, dbDir, port)
	if err != nil {
		return err
	}

	<-ctx.Done()

	fmt.Println("stopping app server")

	return server.Stop()
}
