package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cometbft/cometbft/abci/example/kvstore"
	"github.com/cometbft/cometbft/abci/server"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("specify db dir")
		os.Exit(1)
	}
	dbDir := os.Args[1]
	app := kvstore.NewPersistentApplication(dbDir)

	port := "26658"
	if len(os.Args) > 2 {
		port = os.Args[2]
	}

	laddr := "0.0.0.0:" + port
	server, _ := server.NewServer("tcp://"+laddr, "socket", app)
	if err := server.Start(); err != nil {
		panic(err)
	}

	fmt.Printf("started app: listening on %v / db in %v\n", laddr, dbDir)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// node.NewNodeWithContext(ctx, )

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		cancel()
	}()

	<-ctx.Done()

	if err := server.Stop(); err != nil {
		panic(err)
	}
}
