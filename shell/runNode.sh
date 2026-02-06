set -e
set -x

if [ ! -f /data/config/genesis.json ]; then
    cp -r /shared/* /data/
fi

NODE0_ID=$(cat /shared/nodeID)
echo $NODE0_ID
      
betteryd start --home /data --minimum-gas-prices="0stake" --p2p.persistent_peers="${NODE0_ID}@betterynode0:26656"