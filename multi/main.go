package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/cometbft/cometbft/libs/service"
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
	dbDirs := []string{"n0", "n1", "n2", "n3", "n3x"}
	ports := []string{"26658", "26758", "26858", "26958", "36958"}

	pwd, _ := os.Getwd()

	// if _, errS := os.Stat(filepath.Join(pwd, dbDirs[0])); os.IsNotExist(errS) {
	// 	fmt.Println("testnet setup")
	// 	n := strconv.Itoa(len(dbDirs))
	// 	initCmd := exec.CommandContext(ctx, "cometbft", "testnet", "--v", n, "--o", pwd,
	// 		"--populate-persistent-peers", "--node-dir-prefix", "n")
	// 	if err := initCmd.Run(); err != nil {
	// 		return err
	// 	}
	// }

	fmt.Println("starting app servers")

	var servers []service.Service
	for i := range dbDirs {
		server, err := app.StartAppServer(ctx, dbDirs[i], ports[i])
		if err != nil {
			return err
		}
		servers = append(servers, server)
	}

	time.Sleep(500 * time.Millisecond)

	// start cometbft nodes

	fmt.Println("starting cometbft nodes")

	// node.DefaultNewNode()
	ctx, cancel := context.WithCancel(ctx) // to stop early if one fails to start
	defer cancel()
	var err error
	var cmds []*exec.Cmd
	for i := range dbDirs {
		fullDbDir := filepath.Join(pwd, dbDirs[i])
		cmd := exec.CommandContext(ctx, "cometbft", "start", "--home", fullDbDir)
		fmt.Println(cmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Start()
		if err != nil {
			fmt.Printf("failed to start comet node in %q: %v\n", dbDirs[i], err)
			cancel()
			break
		}
		cmds = append(cmds, cmd)
	}

	// <-ctx.Done()

	// fmt.Println("stopping cometbft nodes")

	for _, cmd := range cmds {
		cmd.Wait()
	}

	fmt.Println("stopping app servers")

	var errs error
	for _, server := range servers {
		errs = errors.Join(errs, server.Stop())
	}

	fmt.Println("app servers stopped")

	return errs
}
