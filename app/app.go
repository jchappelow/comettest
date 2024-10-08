package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cometbft/cometbft/abci/server"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/service"
)

func StartAppServer(ctx context.Context, dbDir, port string) (service.Service, error) {
	// app := kvstore.NewPersistentApplication(filepath.Join(dbDir, "app"))
	app := NewPersistentApplication(filepath.Join(dbDir, "app"))
	app.logger = log.NewTMLogger(os.Stdout)

	laddr := "0.0.0.0:" + port
	server, err := server.NewServer("tcp://"+laddr, "socket", app)
	if err != nil {
		return nil, err
	}

	if err = server.Start(); err != nil {
		return nil, err
	}

	fmt.Printf("started app: listening on %v / db in %v\n", laddr, dbDir)

	return server, nil
}
