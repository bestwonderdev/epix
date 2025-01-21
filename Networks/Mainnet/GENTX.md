# GENTX & HARDFORK INSTRUCTIONS

### Install & Initialize

-   Install epixd binary

-   Initialize epix node directory

```bash
epixd init <node_name> --chain-id epix_epix_1916
```

-   Download the [genesis file](https://github.com/EpixZone/epix/raw/genesis/Networks/Mainnet/genesis.json)

```bash
wget https://github.com/EpixZone/epix/raw/genesis/Networks/Mainnet/genesis.json -b $HOME/.epixd/config
```

### Create & Submit a GENTX file + genesis.json

A GENTX is a genesis transaction that adds a validator node to the genesis file.

```bash
epixd gentx <key_name> <token-amount>aepix --chain-id=epix_epix_1916 --moniker=<your_moniker> --commission-max-change-rate=0.01 --commission-max-rate=0.10 --commission-rate=0.05 --details="<details here>" --security-contact="<email>" --website="<website>"
```

-   Fork [Epix](https://github.com/EpixZone/epix)

-   Copy the contents of `${HOME}/.epixd/config/gentx/gentx-XXXXXXXX.json` to `$HOME/Epix/Mainnet/Gentx/<yourvalidatorname>.json`

-   Create a pull request to the genesis branch of the [repository](https://github.com/EpixZone/epix/Mainnet/gentx)

### Restarting Your Node

You do not need to reinitialize your Epix Node. Basically a hard fork on Cosmos is starting from block 1 with a new genesis file. All your configuration files can stay the same. Steps to ensure a safe restart

1. Backup your data directory.

-   `mkdir $HOME/epix-backup`

-   `cp $HOME/.epixd/data $HOME/epix-backup/`

2. Remove old genesis

-   `rm $HOME/.epixd/genesis.json`

3. Download new genesis

-   `wget`

4. Remove old data

-   `rm -rf $HOME/.epixd/data`

6. Create a new data directory

-   `mkdir $HOME/.epixd/data`

7. copy the contents of the `priv_validator_state.json` file 

-   `nano $HOME/.epixd/data/priv_validator_state.json`

-   Copy the json string and paste into the file
 {
"height": "0",
 "round": 0,
 "step": 0
 }

If you do not reinitialize then your peer id and ip address will remain the same which will prevent you from needing to update your peers list.

8. Download the new binary

```
cd $HOME/Epix
git checkout <branch>
make install
mv $HOME/go/bin/epixd /usr/bin/
```

9. Restart your node

-   `systemctl restart epixd`

## Emergency Reversion

1. Move your backup data directory into your .epixd directory

-   `mv HOME/epix-backup/data $HOME/.epix/`

2. Download the old genesis file

-   `wget https://github.com/EpixZone/epix/raw/main/Mainnet/genesis.json -b $HOME/.epixd/config/`

3. Restart your node

-   `systemctl restart epixd`
