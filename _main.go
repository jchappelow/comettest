package main

import (
	"context"
	"encoding/binary"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/cometbft/cometbft/abci/server"
	"github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
)

type KVStoreApplication struct {
	types.BaseApplication
	store  map[string]string
	logger log.Logger
}

// Initialize a new Key-Value store application
func NewKVStoreApplication() *KVStoreApplication {
	return &KVStoreApplication{
		store:  make(map[string]string),
		logger: log.NewTMLogger(log.NewSyncWriter(os.Stdout)),
	}
}

// Info method returns basic info about the application
func (app *KVStoreApplication) Info(ctx context.Context, req *types.RequestInfo) (*types.ResponseInfo, error) {
	return &types.ResponseInfo{
		Data: "KVStoreApp",
	}, nil
}

// CheckTx verifies the format of the transaction
func (app *KVStoreApplication) CheckTx(ctx context.Context, req *types.RequestCheckTx) (*types.ResponseCheckTx, error) {
	tx := string(req.Tx)
	if !strings.Contains(tx, "=") {
		return &types.ResponseCheckTx{
			Code: 1, // Invalid transaction format
		}, nil
	}
	return &types.ResponseCheckTx{
		Code: 0,
	}, nil
}

func (app *KVStoreApplication) FinalizeBlock(ctx context.Context, req *types.RequestFinalizeBlock) (*types.ResponseFinalizeBlock, error) {
	appHash := binary.BigEndian.AppendUint64(nil, uint64(req.Height))
	return &types.ResponseFinalizeBlock{
		AppHash: appHash,
	}, nil
}

// Query allows to query the current value of a key
func (app *KVStoreApplication) Query(ctx context.Context, req *types.RequestQuery) (*types.ResponseQuery, error) {
	key := string(req.Data)
	value, ok := app.store[key]
	if !ok {
		return &types.ResponseQuery{
			Log: "Key not found",
		}, nil
	}
	return &types.ResponseQuery{
		Log:   "Key found",
		Value: []byte(value),
	}, nil
}

var _ types.Application = (*KVStoreApplication)(nil)

// Commit finalizes the block
func (app *KVStoreApplication) Commit(ctx context.Context, _ *types.RequestCommit) (*types.ResponseCommit, error) {
	return &types.ResponseCommit{
		RetainHeight: 0,
	}, nil
}

func main() {
	app := NewKVStoreApplication()
	server, _ := server.NewServer("tcp://0.0.0.0:26658", "socket", app)
	if err := server.Start(); err != nil {
		panic(err)
	}

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
