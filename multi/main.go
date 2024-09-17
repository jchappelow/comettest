package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
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
	dbDirs := []string{"n3x", "n0", "n2", "n3", "n1"} // n1 last
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

	cmdChan := make(chan *exec.Cmd, len(dbDirs))
	for _, dir := range dbDirs {
		fullDbDir := filepath.Join(pwd, dir)
		cmd := exec.CommandContext(ctx, "cometbft", "start", "--home", fullDbDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		go func() {
			wait := time.Duration(rand.IntN(20000)) * time.Millisecond
			fmt.Printf("waiting to start node %v: %v\n", dir, wait)
			select {
			case <-ctx.Done():
				return
			case <-time.After(wait):
			}
			fmt.Println(cmd)

			err := cmd.Start()
			if err != nil {
				fmt.Printf("failed to start comet node in %q: %v\n", fullDbDir, err)
				cancel()
				return
			}
			fmt.Printf("cometbft %v: %d", dir, cmd.Process.Pid)

			cmdChan <- cmd
		}()
	}

	var cmds []*exec.Cmd
	for {
		select {
		case cmd := <-cmdChan:
			cmds = append(cmds, cmd)
		case <-ctx.Done():
			return ctx.Err()
		}

		if len(cmds) == len(dbDirs) {
			break
		}
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
