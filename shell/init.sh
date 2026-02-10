#!/bin/sh

LOG_FILE=/shared/init.log
exec > "$LOG_FILE" 2>&1

set -e
set -x

CHAIN_ID=bettery-testnet
STAKE=ubet

# 1. Init all nodes
for i in 0 1 2 3 4; do
  betteryd init node$i --chain-id $CHAIN_ID --home /node$i
done

for i in 0 1 2 3 4; do
  sed -i \
    's|laddr = "tcp://127.0.0.1:26657"|laddr = "tcp://0.0.0.0:26657"|' \
    /node$i/config/config.toml
done

# 2. Create keys + add balances to NODE0 genesis
for i in 0 1 2 3 4; do
  betteryd keys add val$i --home /node$i --keyring-backend test
  ADDR=$(betteryd keys show val$i -a --home /node$i --keyring-backend test)
  betteryd genesis add-genesis-account $ADDR 1000000000000$STAKE --home /node0
done

# 3. Copy updated genesis to ALL nodes
for i in 1 2 3 4; do
  cp /node0/config/genesis.json /node$i/config/genesis.json
done

# 4. Generate gentx on EACH node
for i in 0 1 2 3 4; do
  betteryd genesis gentx val$i 1000000000000$STAKE \
    --home /node$i \
    --chain-id $CHAIN_ID \
    --keyring-backend test
done
# 5. Collect gentxs to NODE0  
for i in 1 2 3 4; do
  cp /node$i/config/gentx/*.json /node0/config/gentx/
done

# 6. Apply gentxs
betteryd genesis collect-gentxs --home /node0

# 8. Distribute final genesis
for i in 1 2 3 4; do
  cp /node0/config/genesis.json /node$i/config/genesis.json
done

# 9. Share genesis + node0 ID
cp /node0/config/genesis.json /shared/genesis.json
betteryd tendermint show-node-id --home /node0 > /shared/nodeID

GENTXS=$(jq '.app_state.genutil.gen_txs | length' /node0/config/genesis.json)

if [ "$GENTXS" -eq 0 ]; then
  echo "ERROR: no gentxs in genesis"
  exit 1
fi