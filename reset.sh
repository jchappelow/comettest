rm -r {n0,n1,n2,n3,n3x}/{app,kvstore.db}
rm -r {n0,n1,n2,n3,n3x}/*.db
rm -r {n0,n1,n2,n3,n3x}/*.wal
rm -r {n0,n1,n2,n3,n3x}/data/*.db
echo '{"height": "0","round":0,"step":0}' > n0/data/priv_validator_state.json
echo '{"height": "0","round":0,"step":0}' > n1/data/priv_validator_state.json
echo '{"height": "0","round":0,"step":0}' > n2/data/priv_validator_state.json
echo '{"height": "0","round":0,"step":0}' > n3/data/priv_validator_state.json
echo '{"height": "0","round":0,"step":0}' > n3x/data/priv_validator_state.json

