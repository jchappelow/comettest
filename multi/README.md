## multi node harness

### reset nodes

./reset.sh

### run

go run ./multi/main.go

### reset some nodes

to force a block resync from peers

rm -r {n0,n1,n3}/app {n0,n1,n3}/data/{*.db,cs.wal}

### reset just application data

all:

./app-reset.sh

selective:

rm -r {n0,n1,n3}/app
