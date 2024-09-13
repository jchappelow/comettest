## comettest

cometbft test with dup validator (n1 is bridge) 

### summary

4 unique validators: nodes 0 - 3

node 3 has a clone, 3x

### build

#### build `comettest` (the app)

The ABCI application is separate from cometbft in this project.  Build it:

```sh
go build
```

You now have the `comettest` binary, which runs an app with a persistent kv
store, and a proxy server to which cometbft connects.

#### install `cometbft` (consensus and p2p)

```sh
go install -v github.com/cometbft/cometbft/cmd/cometbft@v0.38.12
```

### run

#### start the app app

In separate terminals, start the ABCI application (separate from cometbft process):

```sh
./comettest n0/app 26658
./comettest n1/app 26758
./comettest n2/app 26858
./comettest n3/app 26958
./comettest n3x/app 36958
```

#### start cometbft

The network is setup with node 1 (`n1`) as the persistent peer that bridges
between two partitions of the network:

1. n0, n1, and n3x (dup of n3)
2. n1, n2, n3

Note that only `n1` is in both groups, so it functions as a sort of bridge
between them. This also allows both `n3` and it's clone `n3x` to connect without
being rejected at the p2p layer.

In this test, we are trying to create duplicate votes from n3 and n3x, causing
misbehavior evidence to be included in a block, and ultimately the validator
being removed from the validator set.

Partition 1 (minus `n1`)

```sh
# run in separate terminals!
cometbft start --home ./n3x
cometbft start --home ./n0
```

Partition 2 (minus `n1`)

```sh
cometbft start --home ./n2
cometbft start --home ./n3
```

`n1`

```sh
cometbft start --home ./n1
```
