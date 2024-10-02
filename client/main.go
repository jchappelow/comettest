package main

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"time"

	"github.com/cometbft/cometbft/libs/rand"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
)

// Tests client behavior: broadcasting transactions and querying tx and app state
func BroadcastTxCommit(ctx context.Context, clt *rpchttp.HTTP) (*coretypes.ResultBroadcastTxCommit, error) {
	// Broadcast a transaction

	// generate random data
	val := rand.Str(8)
	tx := []byte("key=" + val)
	res, err := clt.BroadcastTxCommit(ctx, tx)
	if err != nil {
		return nil, err
	}

	fmt.Println("broadcast tx commit : ", res.Height, res.Hash)
	return res, nil
}

func BroadcastTxSync(ctx context.Context, clt *rpchttp.HTTP) (*coretypes.ResultBroadcastTx, error) {
	// Broadcast a transaction
	val := rand.Str(8)
	tx := []byte("key=" + val)
	res, err := clt.BroadcastTxSync(ctx, tx)
	if err != nil {
		return nil, err
	}

	fmt.Println("broadcast tx sync: ", res.Hash, res.Code, res.Log)
	return res, nil
}

func QueryTx(ctx context.Context, clt *rpchttp.HTTP, hash []byte) {
	for {
		res, err := clt.Tx(ctx, hash, false)
		if err == nil {
			fmt.Println("Tx result: ", res.Height, res.TxResult, res.Tx.String())
			break
		}

		// The transaction could be in mempool.
		limit := math.MaxInt                           // cmt is bugged, -1 doesn't actually work (see rpc/core.validatePerPage and how it goes with 30 instead of no limit)
		unconf, err := clt.UnconfirmedTxs(ctx, &limit) // SLOW quite often!
		if err != nil {
			fmt.Println("error getting unconfirmed txs", err)
			continue
		}

		found := false
		for _, tx := range unconf.Txs {
			if bytes.Equal(tx.Hash(), hash) {
				fmt.Println("found tx in mempool, wait for it to be mined", tx)
				found = true
			}
		}

		if found {
			continue
		}

		fmt.Println("tx not found in mempool or block")
		time.Sleep(50 * time.Millisecond)

	}
}

func main() {
	// StartAppServer(ctx, dbDir, port)
	// BroadcastTx()
	// QueryTx(hash)
	ctx := context.Background()
	clt, err := rpchttp.New("tcp://localhost:26657", "/websocket")
	if err != nil {
		return
	}

	res, err := BroadcastTxCommit(ctx, clt)
	if err != nil {
		return
	}

	QueryTx(ctx, clt, res.Hash)

	// res2, err := BroadcastTxSync(ctx, clt)
	// if err != nil {
	// 	return
	// }

	// QueryTx(ctx, clt, res2.Hash)
}
