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
	"strings"
	"syscall"
	"time"

	"github.com/cometbft/cometbft/libs/service"
	"github.com/jchappelow/comettest/app"
)

var persistentPeers = []string{
	"eb0852d124a184f65b37ff6e9ebcef4818c7eb0b@127.0.0.4:36656", // n0 -> n3x
	"3df4de0b23eb6cd27d806313d6241698a9d953fa@127.0.0.3:26656", // n1 -> n2
	"eb0852d124a184f65b37ff6e9ebcef4818c7eb0b@127.0.0.4:26656", // n2 -> n3
	"3df4de0b23eb6cd27d806313d6241698a9d953fa@127.0.0.3:26656", // n3 -> n2
	"1bb55f8186128b61354d2b1967639cb6ead32757@127.0.0.1:26656", // n3x -> n0
}

var peers = []string{
	"1bb55f8186128b61354d2b1967639cb6ead32757@127.0.0.1:26656",
	"cc4c2c39e90d848b03b46c03d84990912a407e1a@127.0.0.2:26656",
	"3df4de0b23eb6cd27d806313d6241698a9d953fa@127.0.0.3:26656",
	"eb0852d124a184f65b37ff6e9ebcef4818c7eb0b@127.0.0.4:26656",
	"eb0852d124a184f65b37ff6e9ebcef4818c7eb0b@127.0.0.4:36656",
}

func shufflePeers() {
	rand.Shuffle(len(peers), func(i, j int) { peers[i], peers[j] = peers[j], peers[i] })
}

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

type prefixWriter struct {
	prefix string
	writer *os.File
}

func (pw *prefixWriter) Write(b []byte) (int, error) {
	lines := strings.Split(string(b), "\n")
	for i, line := range lines {
		if line == "" && i == len(lines)-1 {
			// Don't write the last empty line caused by split on newline
			break
		}
		// Write the prefixed line to the underlying writer
		_, err := fmt.Fprintf(pw.writer, "%s%s\n", pw.prefix, line)
		if err != nil {
			return 0, err
		}
	}
	return len(b), nil
}

func (p *prefixWriter) Fd() uintptr {
	return p.writer.Fd()
}

func mainCore(ctx context.Context) error {
	dbDirs := []string{"n0", "n2", "n3", "n3x", "n1"}
	ports := []string{"26658", "26858", "26958", "36958", "26758"}

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

	// time.Sleep(500 * time.Millisecond)

	// start cometbft nodes

	fmt.Println("starting cometbft nodes")

	// node.DefaultNewNode()
	ctx, cancel := context.WithCancel(ctx) // to stop early if one fails to start
	defer cancel()

	cmdChan := make(chan *exec.Cmd, len(dbDirs))
	for i, dir := range dbDirs {
		fullDbDir := filepath.Join(pwd, dir)

		// shufflePeers()
		//  pp := strings.Join([]string{persistentPeers[i], peers[0]}, ",")
		// pp := persistentPeers[i]
		// fmt.Printf("node %v persistent peers: %v", dir, pp)
		cmd := exec.CommandContext(ctx, "cometbft", "start", "--home", fullDbDir)
		// "--p2p.persistent_peers", pp)
		cmd.Stdout = os.Stdout // &prefixWriter{prefix: "[" + dir + "] ", writer: os.Stdout}
		cmd.Stderr = os.Stderr // &prefixWriter{prefix: "[" + dir + "] ", writer: os.Stderr}

		go func() {
			wait := time.Duration(rand.IntN(600)) * time.Millisecond // 20000

			if i == len(dbDirs)-1 { // nn2 last
				time.Sleep(2 * time.Second)
			}
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
