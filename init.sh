KEY="mykey"
CHAINID="epix_1917-1"
MONIKER="localtestnet"
KEYRING="test"
KEYALGO="eth_secp256k1"
LOGLEVEL="info"
# to trace evm
#TRACE="--trace"
TRACE=""

# validate dependencies are installed
command -v jq > /dev/null 2>&1 || { echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"; exit 1; }

# Reinstall daemon
rm -rf ~/.epixd*
make install

# Add Go binary path to PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Set client config
epixd config set client chain-id $CHAINID
epixd config set client keyring-backend $KEYRING

# if $KEY exists it should be deleted
epixd keys add $KEY --keyring-backend $KEYRING --algo $KEYALGO

# Set moniker and chain-id for epix (Moniker can be anything, chain-id must be an integer)
epixd init $MONIKER --chain-id $CHAINID

# Change parameter token denominations to aepix
cat $HOME/.epixd/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="aepix"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="aepix"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["coinswap"]["params"]["pool_creation_fee"]["denom"]="aepix"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["coinswap"]["standard_denom"]="aepix"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["gov"]["params"]["min_deposit"][0]["denom"]="aepix"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["gov"]["params"]["expedited_min_deposit"][0]["denom"]="aepix"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["evm"]["params"]["evm_denom"]="aepix"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["inflation"]["params"]["mint_denom"]="aepix"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json

# Set gas limit in genesis
cat $HOME/.epixd/config/genesis.json | jq '.consensus["params"]["block"]["max_gas"]="10000000"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json

# disable produce empty block
if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' 's/create_empty_blocks = true/create_empty_blocks = false/g' $HOME/.epixd/config/config.toml
  else
    sed -i 's/create_empty_blocks = true/create_empty_blocks = false/g' $HOME/.epixd/config/config.toml
fi

if [[ $1 == "pending" ]]; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' 's/create_empty_blocks_interval = "0s"/create_empty_blocks_interval = "30s"/g' $HOME/.epixd/config/config.toml
        sed -i '' 's/timeout_propose = "3s"/timeout_propose = "30s"/g' $HOME/.epixd/config/config.toml
        sed -i '' 's/timeout_propose_delta = "500ms"/timeout_propose_delta = "5s"/g' $HOME/.epixd/config/config.toml
        sed -i '' 's/timeout_prevote = "1s"/timeout_prevote = "10s"/g' $HOME/.epixd/config/config.toml
        sed -i '' 's/timeout_prevote_delta = "500ms"/timeout_prevote_delta = "5s"/g' $HOME/.epixd/config/config.toml
        sed -i '' 's/timeout_precommit = "1s"/timeout_precommit = "10s"/g' $HOME/.epixd/config/config.toml
        sed -i '' 's/timeout_precommit_delta = "500ms"/timeout_precommit_delta = "5s"/g' $HOME/.epixd/config/config.toml
        sed -i '' 's/timeout_commit = "5s"/timeout_commit = "150s"/g' $HOME/.epixd/config/config.toml
        sed -i '' 's/timeout_broadcast_tx_commit = "10s"/timeout_broadcast_tx_commit = "150s"/g' $HOME/.epixd/config/config.toml
    else
        sed -i 's/create_empty_blocks_interval = "0s"/create_empty_blocks_interval = "30s"/g' $HOME/.epixd/config/config.toml
        sed -i 's/timeout_propose = "3s"/timeout_propose = "30s"/g' $HOME/.epixd/config/config.toml
        sed -i 's/timeout_propose_delta = "500ms"/timeout_propose_delta = "5s"/g' $HOME/.epixd/config/config.toml
        sed -i 's/timeout_prevote = "1s"/timeout_prevote = "10s"/g' $HOME/.epixd/config/config.toml
        sed -i 's/timeout_prevote_delta = "500ms"/timeout_prevote_delta = "5s"/g' $HOME/.epixd/config/config.toml
        sed -i 's/timeout_precommit = "1s"/timeout_precommit = "10s"/g' $HOME/.epixd/config/config.toml
        sed -i 's/timeout_precommit_delta = "500ms"/timeout_precommit_delta = "5s"/g' $HOME/.epixd/config/config.toml
        sed -i 's/timeout_commit = "5s"/timeout_commit = "150s"/g' $HOME/.epixd/config/config.toml
        sed -i 's/timeout_broadcast_tx_commit = "10s"/timeout_broadcast_tx_commit = "150s"/g' $HOME/.epixd/config/config.toml
    fi
fi


epixd add-genesis-account $KEY 11844769000000000000000000aepix --keyring-backend $KEYRING

epixd add-genesis-account epix1cttyg9c0e72wcyg4fulxn5pajafvqerfvzg6ta 11844769000000000000000000aepix
# These smaller accounts are adjusted to maintain the target total supply
epixd add-genesis-account epix1shyv0p05z3dhtjr28w37qkcm54yg5ulnhzdwc4 0aepix
epixd add-genesis-account epix1fw4t8peek96a6x6a32v30y22l59ph5wc4hmqmw 0aepix

        # Update total supply with claim values
        validators_supply=$(cat $HOME/.epixd/config/genesis.json | jq -r '.app_state["bank"]["supply"][0]["amount"]')
        total_supply=23689538000000000000000000
cat $HOME/.epixd/config/genesis.json | jq -r --arg total_supply "$total_supply" '.app_state["bank"]["supply"][0]["amount"]=$total_supply' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json

        # Configure community pool
        cat $HOME/.epixd/config/genesis.json | jq '.app_state["distribution"]["fee_pool"]["community_pool"] = []' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json

        # Sign genesis transaction
        epixd gentx $KEY 1000000000000000000aepix --keyring-backend $KEYRING --chain-id $CHAINID
        
        # Collect genesis tx
        epixd collect-gentxs
        
        # Run this to ensure everything worked and that the genesis file is setup correctly
        epixd validate-genesis
        
        if [[ $1 == "pending" ]]; then
          echo "pending mode is on, please wait for the first block committed."
        fi
        
        # Start the node (remove the --pruning=nothing flag if historical queries are not needed)
        epixd start --pruning=nothing $TRACE --log_level $LOGLEVEL --minimum-gas-prices=0.0001aepix --json-rpc.api eth,txpool,personal,net,debug,web3 --rpc.laddr "tcp://0.0.0.0:26657" --api.enable --chain-id $CHAINID
        epixd tx distribution fund-community-pool 23668256824195825000000000aepix --from $KEY --gas-prices 100aepix  --keyring-backend $KEYRING --chain-id $CHAINID
        
# Configure inflation parameters
cat $HOME/.epixd/config/genesis.json | jq '.app_state["inflation"]["params"]["enable_inflation"]=true' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json

# Set first year max rewards (2.9M EPIX)
cat $HOME/.epixd/config/genesis.json | jq '.app_state["inflation"]["params"]["exponential_calculation"]["a"]="2900000.000000000000000000"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json

# Set halving rate (15.9% for 4-year halving)
cat $HOME/.epixd/config/genesis.json | jq '.app_state["inflation"]["params"]["exponential_calculation"]["r"]="0.159000000000000000"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json

# Configure reward distribution (90% contributors, 10% validators)
cat $HOME/.epixd/config/genesis.json | jq '.app_state["inflation"]["params"]["inflation_distribution"]["staking_rewards"]="0.100000000000000000"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["inflation"]["params"]["inflation_distribution"]["community_pool"]="0.900000000000000000"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json

# Set max supply cap (42M EPIX)
cat $HOME/.epixd/config/genesis.json | jq '.app_state["inflation"]["params"]["exponential_calculation"]["c"]="42000000.000000000000000000"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
        