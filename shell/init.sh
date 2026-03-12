#!/bin/sh

LOG_FILE=/shared/init.log
exec > "$LOG_FILE" 2>&1

set -e
set -x

if [ -f /shared/genesis.json ] && [ -f /shared/nodeID ]; then
  echo "Blockchain already initialized."
  exit 0
fi

CHAIN_ID=bettery
STAKE=ubet

# 1. Init all nodes
for i in 0 1; do
  betteryd init node$i --chain-id $CHAIN_ID --home /node$i
done

# 2. Create keys + add balances to NODE0 genesis
for i in 0 1; do
  betteryd keys add val$i --home /node$i --keyring-backend test
  ADDR=$(betteryd keys show val$i -a --home /node$i --keyring-backend test)
  betteryd genesis add-genesis-account $ADDR 1000000000000$STAKE --home /node0

  betteryd keys add faucet --home /node$i --keyring-backend test
  ADDR=$(betteryd keys show faucet -a --home /node$i --keyring-backend test)
  betteryd genesis add-genesis-account $ADDR 100$STAKE --home /node0
done

# 3. Copy updated genesis to ALL nodes
for i in 1; do
  cp /node0/config/genesis.json /node$i/config/genesis.json
done

# 4. Generate gentx on EACH node
for i in 0 1; do
  betteryd genesis gentx val$i 1000000000000$STAKE \
    --home /node$i \
    --chain-id $CHAIN_ID \
    --keyring-backend test
done
# 5. Collect gentxs to NODE0  
for i in 1; do
  cp /node$i/config/gentx/*.json /node0/config/gentx/
done

# 6. Apply gentxs
betteryd genesis collect-gentxs --home /node0

# 8. Distribute final genesis
for i in 1; do
  cp /node0/config/genesis.json /node$i/config/genesis.json
done

# 9. Share genesis + node0 ID
cp /node0/config/genesis.json /shared/genesis.json
betteryd tendermint show-node-id --home /node0 > /shared/nodeID

NODE_ID=$(cat /shared/nodeID)

# FOR HISTORY EVENTS INDEXER
for i in 0 1; do
  sed -i 's|laddr = "tcp://127.0.0.1:26657"|laddr = "tcp://0.0.0.0:26657"|' /node$i/config/config.toml
  sed -i 's|enabled = false|enabled = true|' /node$i/config/app.toml
  sed -i '/\[grpc\]/,/^\[/ s|address = ".*"|address = "0.0.0.0:9090"|' /node$i/config/app.toml
  sed -i 's/cors_allowed_origins = \[\]/cors_allowed_origins = ["*"]/' /node$i/config/config.toml # TODO: ADD ORIGIN FOR FRONTEND
  sed -i 's/cors_allowed_methods = \[\]/cors_allowed_methods = ["HEAD","GET","POST"]/' /node$i/config/config.toml
  sed -i 's/cors_allowed_headers = \[\]/cors_allowed_headers = ["Origin","Accept","Content-Type","X-Requested-With"]/' /node$i/config/config.toml
  sed -i "s/persistent_peers = \"\"/persistent_peers = \"$NODE_ID@betterynode0:26656\"/" /node$i/config/config.toml
  sed -i 's/pruning = "default"/pruning = "custom"/' /node$i/config/app.toml
  sed -i 's/pruning-keep-recent = "0"/pruning-keep-recent = "100"/' /node$i/config/app.toml
  sed -i 's/pruning-interval = "0"/pruning-interval = "50"/' /node$i/config/app.toml
  sed -i 's/index-events = \[\]/index-events = ["message.sender", "message.action", "transfer.recipient"]/' /node$i/config/app.toml
  sed -i 's/tx_index = "null"/tx_index = "kv"/' /node$i/config/config.toml
  sed -i 's/max_num_inbound_peers = 40/max_num_inbound_peers = 20/' /node$i/config/config.toml
done

GENTXS=$(jq '.app_state.genutil.gen_txs | length' /node0/config/genesis.json)

if [ "$GENTXS" -eq 0 ]; then
  echo "ERROR: no gentxs in genesis"
  exit 1
fi