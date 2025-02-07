const fs = require('fs')
const Big = require('big.js')
const axios = require('axios')
const path = require('path')

// Write init.sh in the scripts folder
const OUTPUT_PATH = __dirname + '/init.sh'

const IS_TESTNET = true
const TESTNET_CHAINID = "epix_1917-1"
const MAINNET_CHAINID = "epix_1916-1"
const INITIAL_KEY_AMOUNT = new Big("1") // Small amount for gas
const COMMUNITY_POOL_AMOUNT = new Big("11844769").minus(INITIAL_KEY_AMOUNT) // Community Pool amount in EPIX, minus key amount
const CSV_AMOUNT = new Big("11844769") // Amount to be distributed from CSV
const EXPECTED_SUPPLY = new Big("23689538") // Total supply should be exactly this
const decimal = 18
const KEY = "mykey"
const MONIKER = "localtestnet"
const KEYRING = "test"
const KEYALGO = "eth_secp256k1"
const LOGLEVEL = "info"
const TRACE = ""

const calculateTokenAmount = (amount) => {
    let str = amount.toString()
    let intStr = str.split('.')[0]
    let floatStr = str.split('.')[1] || ""

    intStr = intStr.replace(/^0+/, '');

    let endStr = ''
    for (let i = 0; i < decimal - floatStr.length; i++) {
        endStr += "0"
    }

    if (intStr.length == 0) floatStr = floatStr.replace(/^0+/, '');

    let resultStr = intStr + floatStr + endStr
    return resultStr
}

const generate = async () => {
    try {
        let content = ''

        content += `KEY="mykey"
CHAINID="${IS_TESTNET ? TESTNET_CHAINID : MAINNET_CHAINID}"
MONIKER="${MONIKER}"
KEYRING="${KEYRING}"
KEYALGO="${KEYALGO}"
LOGLEVEL="${LOGLEVEL}"
# to trace evm
#TRACE="--trace"
TRACE="${TRACE}"\n\n`

        content += `# validate dependencies are installed\n`
        content += `command -v jq > /dev/null 2>&1 || { echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"; exit 1; }\n\n`

        content += `# Add Go binary path to PATH\n`
        content += `export PATH=$PATH:$(go env GOPATH)/bin\n\n`

        content += `# Install epixd\n`
        content += `rm -rf ~/.epixd*\n`
        content += `cd ../ && make install\n\n\n\n`

        content += `# Verify epixd is installed\n`
        content += `command -v epixd > /dev/null 2>&1 || { echo >&2 "epixd not installed. Please check the installation."; exit 1; }\n\n`

        content += `# Set client config\n`
        content += `epixd config set client chain-id $CHAINID\n`
        content += `epixd config set client keyring-backend $KEYRING\n\n`

        content += `# if $KEY exists it should be deleted\n`
        content += `epixd keys add $KEY --keyring-backend $KEYRING --algo $KEYALGO\n\n`

        content += `# Set moniker and chain-id for epix (Moniker can be anything, chain-id must be an integer)\n`
        content += `epixd init $MONIKER --chain-id $CHAINID\n\n`

        content += `# Change parameter token denominations to aepix\n`
        content += `cat $HOME/.epixd/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="aepix"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json\n`
        content += `cat $HOME/.epixd/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="aepix"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json\n`
        content += `cat $HOME/.epixd/config/genesis.json | jq '.app_state["coinswap"]["params"]["pool_creation_fee"]["denom"]="aepix"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json\n`
        content += `cat $HOME/.epixd/config/genesis.json | jq '.app_state["coinswap"]["standard_denom"]="aepix"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json\n`
        content += `cat $HOME/.epixd/config/genesis.json | jq '.app_state["gov"]["params"]["min_deposit"][0]["denom"]="aepix"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json\n`
        content += `cat $HOME/.epixd/config/genesis.json | jq '.app_state["gov"]["params"]["expedited_min_deposit"][0]["denom"]="aepix"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json\n`
        content += `cat $HOME/.epixd/config/genesis.json | jq '.app_state["evm"]["params"]["evm_denom"]="aepix"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json\n`
        content += `cat $HOME/.epixd/config/genesis.json | jq '.app_state["inflation"]["params"]["mint_denom"]="aepix"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json\n\n`

        content += `# Set gas limit in genesis\n`
        content += `cat $HOME/.epixd/config/genesis.json | jq '.consensus["params"]["block"]["max_gas"]="10000000"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json\n\n`

        content += `# disable produce empty block\n`
        content += `if [[ "$OSTYPE" == "darwin"* ]]; then\n`
        content += `    sed -i '' 's/create_empty_blocks = true/create_empty_blocks = false/g' $HOME/.epixd/config/config.toml\n`
        content += `  else\n`
        content += `    sed -i 's/create_empty_blocks = true/create_empty_blocks = false/g' $HOME/.epixd/config/config.toml\n`
        content += `fi\n\n`

        content += `if [[ $1 == "pending" ]]; then
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
fi\n\n\n`

        content += `epixd add-genesis-account $KEY ${calculateTokenAmount(INITIAL_KEY_AMOUNT)}aepix --keyring-backend $KEYRING\n\n`
        content += `# These are the airdropped accounts\n\n`

        try {
            const response = await axios.get('https://snapapi.epix.zone/download-csv?detailed=false', {
                headers: {
                    'accept': '*/*'
                }
            });

            console.log('First few lines of CSV response:', response.data.split('\n').slice(0, 3));

            const csvContent = response.data;
            const lines = csvContent.split('\n')

            // Skip header if it exists
            const startIndex = lines[0].toLowerCase().includes('address') ? 1 : 0;

            let csvTotal = new Big(0)
            let seenAddresses = new Set()
            console.log('Expected CSV total:', CSV_AMOUNT.toString())

            for (let i = startIndex; i < lines.length; i++) {
                const line = lines[i].trim()
                if (!line) continue;

                const [account, amountStr] = line.split(",")
                if (!account || !amountStr) {
                    console.warn(`Skipping invalid line ${i + 1}: ${line}`)
                    continue;
                }

                // Check for duplicate addresses
                if (seenAddresses.has(account)) {
                    throw new Error(`ERROR: Duplicate address found on line ${i + 1}: ${account}. Stopping script execution.`)
                }
                seenAddresses.add(account)

                const cleanAmount = amountStr.trim().replace(/[^0-9.]/g, '')
                if (!cleanAmount || isNaN(parseFloat(cleanAmount))) {
                    console.warn(`Skipping invalid amount on line ${i + 1}: ${amountStr}`)
                    continue;
                }

                const amount = new Big(cleanAmount)
                console.log(`Processing account ${account} with amount: ${amount.toString()}`)
                content += `epixd add-genesis-account ${account} ${calculateTokenAmount(amount)}aepix\n`
                csvTotal = csvTotal.plus(amount)
                console.log('Running CSV total:', csvTotal.toString())
            }

            if (!csvTotal.eq(CSV_AMOUNT)) {
                throw new Error(`ERROR: CSV total ${csvTotal.toString()} does not match expected CSV amount ${CSV_AMOUNT.toString()}. Please ensure the total matches exactly.`)
            }
        } catch (error) {
            console.error(error.message || 'Error fetching or processing CSV data')
            if (error.response) {
                console.error('API Response:', error.response.data)
            }
            // Exit the process with error
            process.exit(1)
        }

        content += `\n        # Update total supply with claim values\n`
        content += `        total_supply=${calculateTokenAmount(EXPECTED_SUPPLY)}\n`
        content += `cat $HOME/.epixd/config/genesis.json | jq -r --arg total_supply "$total_supply" '.app_state["bank"]["supply"][0]["amount"]=$total_supply' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json\n`

        // Configure distribution module with community pool
        content += `\n        # Configure distribution module with community pool\n`
        content += `cat $HOME/.epixd/config/genesis.json | jq '.app_state["distribution"]["params"]["community_tax"]="0.000000000000000000"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json\n`
        content += `cat $HOME/.epixd/config/genesis.json | jq '.app_state["distribution"]["fee_pool"]["community_pool"] = [{"denom":"aepix","amount":"${calculateTokenAmount(COMMUNITY_POOL_AMOUNT)}"}]' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json\n`
        content += `cat $HOME/.epixd/config/genesis.json | jq '.app_state["bank"]["balances"] += [{"address":"epix1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8j52fwy","coins":[{"denom":"aepix","amount":"${calculateTokenAmount(COMMUNITY_POOL_AMOUNT)}"}]}]' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json\n`

        // Configure inflation parameters
        content += `
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
        cat $HOME/.epixd/config/genesis.json | jq '.app_state["inflation"]["params"]["exponential_calculation"]["c"]="42000000.000000000000000000"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json\n\n`

        content += `
        # Sign genesis transaction
        epixd gentx $KEY ${1 * 10 ** decimal}aepix --keyring-backend $KEYRING --chain-id $CHAINID
        `

        content += `
        # Collect genesis tx
        epixd collect-gentxs
        `

        content += `
        # Run this to ensure everything worked and that the genesis file is setup correctly
        epixd validate-genesis
        `

        content += `
        if [[ $1 == "pending" ]]; then
          echo "pending mode is on, please wait for the first block committed."
        fi
        `

        content += `
        # Start the node (remove the --pruning=nothing flag if historical queries are not needed)
        epixd start --pruning=nothing $TRACE --log_level $LOGLEVEL --minimum-gas-prices=0.0001aepix --json-rpc.api eth,txpool,personal,net,debug,web3 --rpc.laddr "tcp://0.0.0.0:26657" --api.enable --chain-id $CHAINID
        `

        fs.writeFileSync(OUTPUT_PATH, content, 'utf-8')
        // Make the file executable
        fs.chmodSync(OUTPUT_PATH, '755')
        console.log(`Generated init.sh at: ${OUTPUT_PATH}`)
    } catch (error) {
        console.error('Error generating script:', error)
    }
}

// Execute the script
generate()
