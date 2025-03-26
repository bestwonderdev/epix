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

# Add Go binary path to PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Install epixd
rm -rf ~/.epixd*
make -C $(dirname $(dirname $0)) install



# Verify epixd is installed
command -v epixd > /dev/null 2>&1 || { echo >&2 "epixd not installed. Please check the installation."; exit 1; }

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


epixd add-genesis-account $KEY 1000100000000000000aepix --keyring-backend $KEYRING

# These are the airdropped accounts

epixd add-genesis-account epix1jaw0pfutjgz5t3pkzws6k58u29u2flqv6mt32c 272982518775450000000000aepix
epixd add-genesis-account epix10y2jd8mcveelz27kc3jdyggndlgg9nqg2te8ts 19599570183580000000000aepix
epixd add-genesis-account epix1marvg095283tnkta9p2x5patck93dlm6e6aqwu 1676302353000000000000aepix
epixd add-genesis-account epix1eznqnlsqda43kj022qkp4enw8486srw53e2ldq 9908440641000000000000aepix
epixd add-genesis-account epix1fu9eq5qzlemrmjw0c3vrq37yxmv8hut8t98jxf 14481642844310000000000aepix
epixd add-genesis-account epix1xu2jufpfxc9ajtejje5c7vdmn6ddtfrmrwcxj6 87404592331190000000000aepix
epixd add-genesis-account epix1mrdzx6ecv45y5crz2xxxzx6dthz56m0t9mee5c 22427511580190000000000aepix
epixd add-genesis-account epix1q0px6h3zcrv4pq5peh7dsk2dmdggasehznpccm 35872870603150000000000aepix
epixd add-genesis-account epix1a54e5t0x7rqyj78wp5ra272z8v8uaa263x5jfs 187734176860000000000aepix
epixd add-genesis-account epix1d9768xl5u84ejex57x9lhagg3paakp5eynzs0x 3900180297590000000000aepix
epixd add-genesis-account epix1dz76dqr6mzqve7cyax8q8jgnr9klm8p3lqw5pl 19887843905060000000000aepix
epixd add-genesis-account epix1qpggj3adws5p3f97u2frnf9zh83gyws3mlljah 21755433650640000000000aepix
epixd add-genesis-account epix1ggsmk2fjjw476k0urxd3g4et67f267eknsd076 3167855734000000000000aepix
epixd add-genesis-account epix1mkum8hhzk2v85nnljc88hxxga5muvwlseryvaq 16773582762170000000000aepix
epixd add-genesis-account epix1uv8py5rxvgqkxcpvrtqxusmn34v32amtdmz75x 20536133609710000000000aepix
epixd add-genesis-account epix1fchavq74lnxt72he263s9367t3n04fuzu8q347 1807902949660000000000aepix
epixd add-genesis-account epix1n4m2qt02z7hwqqsj0pllsqcc4np9c423glkcyd 22721700376610000000000aepix
epixd add-genesis-account epix17jtzy4j5dzq225r8m96vd386n4jgewp82ql24x 20101681696190000000000aepix
epixd add-genesis-account epix1s2dh6edze04sldk06ukmz6whzdpj3ghgquzckr 1791522243910000000000aepix
epixd add-genesis-account epix1p4nhs45alrmmmkgz6yv6h2cs0m89ulvp352a00 20934317777400000000000aepix
epixd add-genesis-account epix1v8wyvccvchd638cd4rh4c0p4q2ru6kmxx9yrw4 5602117729670000000000aepix
epixd add-genesis-account epix1m2gns704fhe2vx9e9r6aqluwsp7zg5zmd6kafc 17575699773100000000000aepix
epixd add-genesis-account epix1ayyrjhr5d865k7qt0w38pqs4pwk3k7luw9qppr 14685867253970000000000aepix
epixd add-genesis-account epix1sulswqvrzy63489p4ltclgavn5cu9kgkwrtqf7 22748314783720000000000aepix
epixd add-genesis-account epix1pym5fadujngxnfw9s98nxspyzdve9kvlyw655h 20758190589680000000000aepix
epixd add-genesis-account epix1heun4fq6kcfk49y3urgxwfxeke8espd9rq8xy3 17766425589840000000000aepix
epixd add-genesis-account epix1n5w20ycluew5zz2jrzxta4nfenu340u5tc2anv 20445158975680000000000aepix
epixd add-genesis-account epix1e492te60hthjj6p0dpjqggq9trljaunj9d73vj 20583220386430000000000aepix
epixd add-genesis-account epix193vf2h9nzt7akhq42qv8p0fqxkkac3ut2daerd 18433975411570000000000aepix
epixd add-genesis-account epix173rpuaruvnz8ksv7m6xgkdhuya3960f7qk9k5s 1514433164550000000000aepix
epixd add-genesis-account epix1vjuew9nc4nu9fkg7tncvsfelfyy63w6stkz2xc 14424032601490000000000aepix
epixd add-genesis-account epix1787sc2cmxad4sjm00s3jx0mgcjujfuhuaeky7f 21520589302140000000000aepix
epixd add-genesis-account epix1np8xh0a2j5rdm8gdjk7vn6ngc46984u8gp4ffz 3894222142430000000000aepix
epixd add-genesis-account epix1javwveqmfgduwpvm9cxsyvr6npmt5w6yzpp8kn 7813878319330000000000aepix
epixd add-genesis-account epix1dzsgznqhnqrhpsrjzj7xvjdxaukz7uahclyj55 18894210439430000000000aepix
epixd add-genesis-account epix19gtmgvs2v8lqw4egefezpt5934qpmmr48p6vnv 22333531451360000000000aepix
epixd add-genesis-account epix1p9s83lhjjhujrw4ncvp0w8yyrxsn9prfvasek5 7540536181650000000000aepix
epixd add-genesis-account epix1htf4fnl5lf4slmp52pjasx8y048wy8regfptyj 16897163641100000000000aepix
epixd add-genesis-account epix16hxgjscaypagke3q03h6g3jszpwafc9c7fgu3w 17685610015190000000000aepix
epixd add-genesis-account epix1n8at58cqf77937fzqw0em7fa5mu4mws9w8sqj2 3984487096480000000000aepix
epixd add-genesis-account epix1p9z4tz0ehgqtp4umw8le4w0euatjkstpltfcsh 21500142664080000000000aepix
epixd add-genesis-account epix1rfqu6yp4wyd9pn5lqk3grxx0kfuttskuzymq7p 10016580879720000000000aepix
epixd add-genesis-account epix12slphzsdhf4zu6mxwzjhdgvn6au94v9sezau23 19621071573690000000000aepix
epixd add-genesis-account epix1aqxzytfe082c2ecrsd825qntejaajd2t7t7e3f 19890562828610000000000aepix
epixd add-genesis-account epix1dwj30lu3vuwzkr9pkjaku6q97hm2pl06n6fjs0 1813893697780000000000aepix
epixd add-genesis-account epix18gkkh3dsy35zwfw3dwute6px8w9dzj59phvsyp 21471782256360000000000aepix
epixd add-genesis-account epix15ele5jtk3z72prz6lkz6uqk0c89j2v57c0ucqa 22607861880620000000000aepix
epixd add-genesis-account epix19fv0xqx8yh7gdsw9cxawcc0gwl602pkx7e5v04 9652532037280000000000aepix
epixd add-genesis-account epix1ff083apgzh35kk6p4ydt7662u3k4tr6d7hhjhw 14136902679840000000000aepix
epixd add-genesis-account epix18x708nkq3tnuwfdu3kuftkz3lgg88frns7vlm5 22381538408420000000000aepix
epixd add-genesis-account epix17wmnqe8gmslkee2up4rzle0mjkvs5xnvtzs8zf 20763490742700000000000aepix
epixd add-genesis-account epix1876e20yv5jn86ynpazuj22v36fflncpjw20f8n 22102195801710000000000aepix
epixd add-genesis-account epix1xuv0qd3pn7k6j2hu4p88c4w2mf8xag6f0d8ah5 21211549929160000000000aepix
epixd add-genesis-account epix14jq9ly9q8ewwu49hm8hkxd3m8548y7qrsyp0ru 18664704944150000000000aepix
epixd add-genesis-account epix16az7mnsen9yy5c4pty5qfdlehlpqehjzkemw33 5646910070930000000000aepix
epixd add-genesis-account epix1syw3mjezlpf6uzm0vh8an054exdhap5v4k0xrl 16591447353630000000000aepix
epixd add-genesis-account epix1jvzee79rsx2h4sdeg2hsyjfcsk4urnqz07k5tg 1709250168410000000000aepix
epixd add-genesis-account epix1227dpxm6n9vvx743afc92zshq0dgn9lyk5cufh 20353943493100000000000aepix
epixd add-genesis-account epix10uu6sgca9fmwaaz60gw079y5l5rmntrphv2cut 22667013705330000000000aepix
epixd add-genesis-account epix1de35fr0h4uw7unmuvjne54zqp9udjmwpgj4686 16845455047990000000000aepix
epixd add-genesis-account epix1jme5ruwgd3m9s3emaln56f38d6k5ka75s2ft66 16385461179530000000000aepix
epixd add-genesis-account epix10r7lkf5eru0wnj839dl0rf2dnuh3jdksvsd4t3 20261639926370000000000aepix
epixd add-genesis-account epix1tqpl5q9x8xsna6epp2h8zy7f4nfncnwu4z4vla 5848977644380000000000aepix
epixd add-genesis-account epix1h0ewx9sz8z0wluqyz4lazjn56v8lde7pu2psz9 12056998748180000000000aepix
epixd add-genesis-account epix1p7k6c8vlafyc86y2dtc7cru9ntvr92muguk06k 19392968990770000000000aepix
epixd add-genesis-account epix1309n6jtdgqm7y8dtg6c3f8uq84ru2z7l8ch0xl 20939554020880000000000aepix
epixd add-genesis-account epix1utvwc0zw40jhwa6dqtwmgt2gumaytz2ye2hrv2 18282398440670000000000aepix
epixd add-genesis-account epix18dr6f90g3yptqnnkf6y0za9r8frv9ykv6maky2 17161559962040000000000aepix
epixd add-genesis-account epix1z0lur0dr0e72duj7w58sqce8e2fewdsr736cnk 21864477563680000000000aepix
epixd add-genesis-account epix1j5cf9qws9arx52hez8pvzwxt8kk5573a6y4745 1824216744680000000000aepix
epixd add-genesis-account epix16frfuwsh0teec32r5jefj3rx0qh9fxf6g3rttc 22531819061010000000000aepix
epixd add-genesis-account epix1de8z73wuww2dxjnt76mv6tm2n4wn0xssl5sczx 23347925887620000000000aepix
epixd add-genesis-account epix10587a6nv6x6m0xyn76w27avwzng50lw85m9hu6 20744287585490000000000aepix
epixd add-genesis-account epix1ad2jdur7pgnwwtz0q8vqhjkh9uc7dw8qcwha3h 21723588182180000000000aepix
epixd add-genesis-account epix1xcdzapnymxcme33ddlfzl8amp386x0lgl7wrwr 20205615734790000000000aepix
epixd add-genesis-account epix13lm5a997u6akcmketvaddpxd6fgu5cqdxgghjh 21749907812340000000000aepix
epixd add-genesis-account epix1rlz6x6ajjdjf8zucks6w33az4pdjygpukyplnn 3656863328890000000000aepix
epixd add-genesis-account epix1fr2x3gw4xur7mcvtgxf9q6fym97ycn7ths9def 21810587069860000000000aepix
epixd add-genesis-account epix1sfa4w0mquyxpu309y6yep5yny2p40t7w4yqu9c 17467632231610000000000aepix
epixd add-genesis-account epix14q4cawxeac4dcycmxkwrk078gnx0f5a4zgyejj 1820103124010000000000aepix
epixd add-genesis-account epix1p7n3qvcztwmxevzyglrmt0qmq4yj797stsdktw 23198039434200000000000aepix
epixd add-genesis-account epix14z6mxf67g9kgjtuqghhh8qu9cpslcjv58jwp7y 17454214769010000000000aepix
epixd add-genesis-account epix1f9kj99ntp3a2uy2f4u4ns3jceuk8xgrwg9m03q 21276693200340000000000aepix
epixd add-genesis-account epix1x5xtnqedyuk70c5tnh05ucrracmw2zg4wjzkpy 20211728692990000000000aepix
epixd add-genesis-account epix134q4ylx7m4vd77fz75eq8xwrdqffzwtm4n3sp7 3544240146940000000000aepix
epixd add-genesis-account epix1435ql43chh69s7drm9usgsu7zl4un7cwwvc5m9 21399091680320000000000aepix
epixd add-genesis-account epix17cyav8mg58x0pyssr2tj43wd599fyxau5czrjw 18719579251430000000000aepix
epixd add-genesis-account epix1s5ufw0rtkcjr5n5c6xtv8lgf42fxzwxcz6q079 14764689629210000000000aepix
epixd add-genesis-account epix1rvgyw8gjdd5cs2eulqe8zpw7m9slrz8l0zy5nl 19359605381710000000000aepix
epixd add-genesis-account epix1glep40kew0vfc7dlgw9r49wf7dvvr5a3qn8tlm 21955106229630000000000aepix
epixd add-genesis-account epix15nwxq4lmzefu40adqclm5ywtz2mjrv5d9fc226 16037646794430000000000aepix
epixd add-genesis-account epix1lgn2qkwuq08wx5vu8vjghkg5a4lyltwx6gtux5 23231987329520000000000aepix
epixd add-genesis-account epix1hy0frh7hevmq95zxl87ph46cweveyst8x7uy6f 18044895886040000000000aepix
epixd add-genesis-account epix1pgws545ghm6g4nfp0pvpccjjdjee3u8z8w03h5 12334776357810000000000aepix
epixd add-genesis-account epix1pzhntlp9anqrn60jkkfl8qtwa6ngk6ryls5l2s 17409643301290000000000aepix
epixd add-genesis-account epix1lw4rp3djsg3rcnq69vmje7a0578hzu92cvxm6e 21362201359270000000000aepix
epixd add-genesis-account epix1uxs90trt4aeqye6dzt53979g9jm9cy58xspz5g 17422697157600000000000aepix
epixd add-genesis-account epix1q4nhfvg2dhs995wzutyjlq943j0aja2c04llkh 18859572396320000000000aepix
epixd add-genesis-account epix1pfnsa6h3f828r8llgtvyj2smed5ckf0lq6v87f 22367557163650000000000aepix
epixd add-genesis-account epix1fexhw5xwfyu460arzfmd0mg0cll9gnh9t859ru 18569279925930000000000aepix
epixd add-genesis-account epix1ngzna3lkyuyjluqhufzud7kx489u0d25u7ax3q 3856275147390000000000aepix
epixd add-genesis-account epix1tv9x7xz4ss9fu589gw8jaj6kpwkk68sf9dmpu0 5743708546500000000000aepix
epixd add-genesis-account epix1rhwehpd38rftnuc2wuxtwknvl9un0m9dp3cu9y 20840009707120000000000aepix
epixd add-genesis-account epix1s3n3072kjey8qx0dpe4m4lu4ckdh253pm5urgk 19577245639850000000000aepix
epixd add-genesis-account epix1und7qet67h3nzyrhy5wva65m7wtvv5cdf6tdze 18031179312960000000000aepix
epixd add-genesis-account epix12dlffdthum8v6nayk9qzyua4v5k70n56rrr8pl 23452894147960000000000aepix
epixd add-genesis-account epix1ndz3f6elwsj9jl5cfe3hkvfptcfy99l5h22th3 1837473540470000000000aepix
epixd add-genesis-account epix1e0ymauwnrjnkwa0yl6y5d93qkejd3xl6uyw4ph 17306646536290000000000aepix
epixd add-genesis-account epix1m236g3nyaxuhh2w3xs77jt6qmdfxlwnh28wvuq 7982718281900000000000aepix
epixd add-genesis-account epix1jjggzpthgjxzwcuq7tskjpgkektxa742elpjc7 22903772703480000000000aepix
epixd add-genesis-account epix12dceqfx9cdyj2vahs2amtwxskeur0llg52dh80 1866062624820000000000aepix
epixd add-genesis-account epix1ntwyvxfqez4whkjyp6hqzlef36q2j4lu605d0r 21042143635370000000000aepix
epixd add-genesis-account epix142kker4d2lusx22rzm7aa0pvft42tx3t4ek6nh 19885690274640000000000aepix
epixd add-genesis-account epix19h5sxmrf5l692k7yz09wky4rupy7znckm5l64m 21932136422720000000000aepix
epixd add-genesis-account epix12l80kck745fttskuke0tfp8dcekzczfepqkjsg 17592697860510000000000aepix
epixd add-genesis-account epix1dkjwxev7fehfctgym4zhjfdggyrc2dey9ehvne 19731077482730000000000aepix
epixd add-genesis-account epix1gc6j9q7ydeevsmzqr08kru0ucrfvgs758fhqrw 3939537088310000000000aepix
epixd add-genesis-account epix1l6hxfvs97hdtsuxsamt2py3ahwwkh2md2amc7d 21068353919900000000000aepix
epixd add-genesis-account epix1eel759ecdd7y625geeetdx4py3xccldnnp7s47 23300633643320000000000aepix
epixd add-genesis-account epix1edtzmjnecard8uwzjg98td4w0dqdw26zlf3fkq 1866947758290000000000aepix
epixd add-genesis-account epix1gl87unmdqq0f6s430t9xjd5zuc0n2gegzcw60n 18274460035720000000000aepix
epixd add-genesis-account epix1jd3mlzkprl4naehhgqr4gxkt0usa9sa7ca0vly 17211869645790000000000aepix
epixd add-genesis-account epix1at5hdxfxdf72xk7lrspq6dd5d0cqy5edmme0jn 18873170479090000000000aepix
epixd add-genesis-account epix1dc724qafgz5s8dup4kjm6qnj6cyr9xrka98x5h 22091215246090000000000aepix
epixd add-genesis-account epix1upsxlgxm0fs22xexry2ghyncsn9q3pdtu7ztx7 21896852274480000000000aepix
epixd add-genesis-account epix10749wnh4vcghxjvpuvmjkul4gm7q3hk8vef9j3 18695610098630000000000aepix
epixd add-genesis-account epix1y6v40k2k7ky207f28d2f3eqwxsclfwelpuqrw5 21738546941710000000000aepix
epixd add-genesis-account epix1ff4sx24wzxnfnra8rc3c6gmucdf8yy6cz9q43v 20639300298550000000000aepix
epixd add-genesis-account epix120m4tkeh78cfy7nphdvc5kqtkpx9fvkd0g3hwp 1687973964450000000000aepix
epixd add-genesis-account epix1g0urymru4rvflzpwxg2q8gh4wnwp2vs94dyxjq 20463303496410000000000aepix
epixd add-genesis-account epix1djxj3zew5f03ng0yj5ns7h3nku8fsyn4adkemh 17241825675580000000000aepix
epixd add-genesis-account epix1qzmrnsjtrx443w6pauen23wkfqysmp6vx3ltng 17507824992160000000000aepix
epixd add-genesis-account epix13l6fst34vhdg0n7ma7clc52pktyadnxpwsy88k 20667035953940000000000aepix
epixd add-genesis-account epix10h9zla6vhcjf5pn0dg6s58830vu3jxjujehwhy 17151931341780000000000aepix
epixd add-genesis-account epix1upjhk0e4cqrltzgpujkaetlraxxxvy26c90prc 5809033660620000000000aepix
epixd add-genesis-account epix1jngtdxregxh6k0ge62n9g92ud8xefq5wa6ufw7 3838339998200000000000aepix
epixd add-genesis-account epix1y6d3q0y7y6yxuzm0hz672qf2y3zvpuvmhfml9p 20950398097020000000000aepix
epixd add-genesis-account epix1789c8jk4ey3sl9f60j2ddk6cwxk6pc54s7all6 5839118576940000000000aepix
epixd add-genesis-account epix16zglyx7qv9x46x2aep9qkumrf0ykrck0qkjzhy 17387766599720000000000aepix
epixd add-genesis-account epix1uzspne7vfzluuq6206xylj4fktz4v75aw09pz3 20098371402940000000000aepix
epixd add-genesis-account epix1n5z07cpdj4gx7tq5pym8p7cak4cl3l9zlvxxvn 20091284685010000000000aepix
epixd add-genesis-account epix1nyrg750qm3jc7u9qcxte4hverd3qz27uqh6cd0 19795858407540000000000aepix
epixd add-genesis-account epix1vgrvuqu29cn2ye3nuvmly2mtga724hp3ekdwpj 17780157192130000000000aepix
epixd add-genesis-account epix17acllns84veuqf29sshe88c9kjjkl6plnz3mu7 3822408670590000000000aepix
epixd add-genesis-account epix1avjpnk3pmkvnxwvht5ttkd3dmplfn9hft7stgh 3819337027040000000000aepix
epixd add-genesis-account epix150wex5csx2kvhmp8s3yya4arlm7rkpumjzcsa6 22748920523520000000000aepix
epixd add-genesis-account epix1jz4q5lyy5yvap6jcm45asx7h4fwzhawzflkyag 16214917783460000000000aepix
epixd add-genesis-account epix1jgjnwmggx4anepn9jnj0luk2ud3ss5ugdvsv2a 17371797187580000000000aepix
epixd add-genesis-account epix1zs3he2vf5mcl62ceej84cw7jv7tcz5zsghqmpt 5927794722170000000000aepix
epixd add-genesis-account epix1yqz45w5c4vm8zgtlphf5l30d08jk434l6l76hh 21082496446200000000000aepix
epixd add-genesis-account epix1ey553wfg2xruzq4sayd9wlqyafy9kahzgsd6pr 325302238800000000000aepix
epixd add-genesis-account epix1c4chxnn48mzu2m73lgdxyt3cgsmaldrklvdjdt 7059800242880000000000aepix
epixd add-genesis-account epix1249apkh8h4an55szpcm5gpz8wplac3j7mnkk5w 76692744717830000000000aepix
epixd add-genesis-account epix1nac9seck74qgghg78rzluxv33wuq64nlt60yj7 34107501347300000000000aepix
epixd add-genesis-account epix1un84fl7uss65szqa2amn6cssprqwvc999mw0j9 5363571154610000000000aepix
epixd add-genesis-account epix10gu7ayusxaqtnkd9e28y6n3wq5dm7tc9mfa6eu 24457991627710000000000aepix
epixd add-genesis-account epix12nd40dwreyl84tcpv8t53n7uva9rnwm6hq5ygs 47519428574180000000000aepix
epixd add-genesis-account epix18a6hh5n4mpxjvkfh9khs8nwr8ttrd3sk7vhd9t 33897829345720000000000aepix
epixd add-genesis-account epix1aezh664ph36ryqkk85r3hnuj8nsjyzujzslj0a 18838937526300000000000aepix
epixd add-genesis-account epix18h63mrpp3yeackt0t3pqt8uyh6mlnf9p7jpaq5 444002110368870000000000aepix
epixd add-genesis-account epix1zaewdesj75czagec2p89k3qss6h3dxy3j0djj8 379949415944540000000000aepix
epixd add-genesis-account epix17d06xtttq98ccghkl093n6ak680vamker9f5r2 20840683235380000000000aepix
epixd add-genesis-account epix1tleyqh07trj5sv27kvxj5fw40pvxz207qttyvc 2882498050790000000000aepix
epixd add-genesis-account epix1z9zmm2kqcp3vtq5zfasjlsep4d3almazz8mxzv 1735656730000000000aepix
epixd add-genesis-account epix1vd0fmqsenp4s4ykcxghj2q5c7g572hkyt5wptg 858808095800000000000aepix
epixd add-genesis-account epix1mp5sac28l34au5c8u03pzqxxz97l8zr2n8wr8t 723669864520000000000aepix
epixd add-genesis-account epix1rxupq9pc6d79nv3wsnjmplqcevkuh6kj74nssv 169551808270000000000aepix
epixd add-genesis-account epix1k5seqkmfakhvtednz4pqseswtmlsvl3wc55h34 1122565842240000000000aepix
epixd add-genesis-account epix1g4hkcmtrj3jxp74yr8vtmhfn9pva3ux0dvj9g3 957222720100000000000aepix
epixd add-genesis-account epix14xgzq45vc8wadxfx649g9pwhsuw4xa6vk25qef 1017751187250000000000aepix
epixd add-genesis-account epix153jt8r0vp2w04qe8zdv6h62gxx8fk3uzrjqcsc 32777421438170000000000aepix
epixd add-genesis-account epix1mfv5vtjnevc3nqcrsn7v257vstddqtdk0hqk75 297951125308650000000000aepix
epixd add-genesis-account epix1n74w5r3r9g9ap9tgzl843z0vuf9nrjhm5q7cpy 796995916067420000000000aepix
epixd add-genesis-account epix18eqnu246knhz4udfd5ud0ryz6xnd7fvp3syps4 296466518917800000000000aepix
epixd add-genesis-account epix1s89j4wqr739m7vnv7kpug72f3xpzdfgfxq5y0j 51450864300000000000aepix
epixd add-genesis-account epix1tu8vnafhrrqj776k27czkpgfwmqv0n7sqendkf 829852650000000000aepix
epixd add-genesis-account epix1ka5nmpxj384zksgpe6l6geduf4sged0v5fjrp8 2489557950000000000aepix
epixd add-genesis-account epix1z9afgqgzyvfly0tnupwygs3jl00k802wulyvvc 26532697896860000000000aepix
epixd add-genesis-account epix166qh9f57fcyaqtz80jwj84lt9xvxdqu8u57rcj 224669926674700000000000aepix
epixd add-genesis-account epix192gdwypp58f46ua3uugzxx9mpwjtd62wun4am6 44913285123300000000000aepix
epixd add-genesis-account epix1eyp44l3syhypyr5wsgxcj3jxppp86zm8slutjh 44956437461100000000000aepix
epixd add-genesis-account epix16w4d808dv3xdd9yw0qv72hnvdkg2x9xxg7d2p8 44908306007400000000000aepix
epixd add-genesis-account epix148lvwkdc8v5n330wu74j5htezq5mn9cyyz9m6d 15702264380130000000000aepix
epixd add-genesis-account epix17jmahye5dghh0ve2w9szafz085lpuzcsrqk8y2 26917100555400000000000aepix
epixd add-genesis-account epix1e3ywv6kcasuh35cp7p64xdljars45ywp6v0ndd 26956933482600000000000aepix
epixd add-genesis-account epix1rm7t4erl5sz6uj70prt7tg27vh0pes9flgw609 15798527287530000000000aepix
epixd add-genesis-account epix1a4r37vh4saf3vxvu26kdug4uc738080z6lmwka 445610956586400000000000aepix
epixd add-genesis-account epix1jt8jgu4jmqevu3ks9x7zsmae7284v7fkauxxwp 169859219517900000000000aepix
epixd add-genesis-account epix1wf798updnvdca8u4zxjh4c0gc3wtnev9hjmkh7 132451484821730000000000aepix
epixd add-genesis-account epix12y0jfks3lw68m6rd3nkkpu7pv4wwda9pn9mmll 148234763566820000000000aepix
epixd add-genesis-account epix12395cyxstf4yrrvw08c9kpugmgs63nehgj0wgx 104944203434810000000000aepix
epixd add-genesis-account epix1ns4kkslamsyf0r8sf93l4e8j6rehe302etwkxv 65179479822090000000000aepix
epixd add-genesis-account epix1n3q38u6xenfank44zc08ky7a4mmgpgr8n705kk 141641065604670000000000aepix
epixd add-genesis-account epix17nhmsz4yw44zw7nx5zty6a9fyr25xxp8mxtk2z 62851483737280000000000aepix
epixd add-genesis-account epix12qrv0w9rfqxy6vcr6xxmpwpq786r2a6mxdxgkh 47940587590500000000000aepix
epixd add-genesis-account epix15t7ruz25ju347k7lfhrl7eg4d6s9vgzwgqny2d 67249184124370000000000aepix
epixd add-genesis-account epix178l7xm7tvd74ztv7wzyxcc9wda3favj989z0zn 2157616890000000000000aepix
epixd add-genesis-account epix13m3d5llvh7jsf8n9swcql0srp5tcpjlfzfxu3a 2075461477650000000000aepix
epixd add-genesis-account epix1snmgvph50wtlammnlluuchmv0ex39zzfeg2rue 2082930151500000000000aepix
epixd add-genesis-account epix1ufd4we0jc949rxreeqx2k26t7xz6vfm5ylkrf0 2082930151500000000000aepix
epixd add-genesis-account epix1h2ze567fskevkw3nfscfa5re4x2324fufz00gr 2082930151500000000000aepix
epixd add-genesis-account epix1j34jz9xu6e3mh6qzh4pcnqgkvt54jp76u2r99j 2074631625000000000000aepix
epixd add-genesis-account epix14px70yk58m33n0nd65s6nzng5ce27zjwkxjtws 2074631625000000000000aepix
epixd add-genesis-account epix1st25q23c4n85a0fr0fpwp7nmr3pl2ke6qkw4ky 2074631625000000000000aepix
epixd add-genesis-account epix1gf82cpm4vq5z2hxup5qtum9rwm7zznryn2rsvs 2074631625000000000000aepix
epixd add-genesis-account epix1warh9mm96fqv0ze52x5kyxn7mm6vpj0hwdgfza 2074631625000000000000aepix
epixd add-genesis-account epix1g3568mc0qfsx23mm00y5pkayxuv9jxgjhwg9hh 2074631625000000000000aepix
epixd add-genesis-account epix15qq47apnvscrmmf49g2h2a5aehd0wcs3vukc3t 2074631625000000000000aepix
epixd add-genesis-account epix1pn00t8ss748c8skemege04p7fts22jssy7h2eg 777764697010000000000aepix
epixd add-genesis-account epix1064gwcnjcdrj0knqxr7kxu9rswmznddqxr8vc8 829852650000000000aepix
epixd add-genesis-account epix1w6n8zcxjdvjj6f62ytm5dvz5pqxr90qpw33ant 35246364670000000000aepix
epixd add-genesis-account epix17fnujvcquj2qjqzvtjzeej0y4vm470et5q4zma 726017270000000000aepix
epixd add-genesis-account epix19krjhdwrvw0w04ems7dyz8mk8t7g7x937n3v3p 8298526500000000000aepix
epixd add-genesis-account epix1s5d5h3l28xpkxhqn4wslstn9lls7zxxeaapr5z 4149263250000000000000aepix
epixd add-genesis-account epix1mdl5zvn20h3s6w8aa8ss75u578mcr2g5pqxgww 4149263250000000000000aepix
epixd add-genesis-account epix12ta0937gdkw9q66jfw6yl9r53gsxe4hfcad7ya 4149263250000000000000aepix
epixd add-genesis-account epix18aqh8ynzywzpfqvhrt0c9j42sm4w899rmefxq9 162651119400000000000aepix
epixd add-genesis-account epix18t0qy4ftmgssqatalmre7lun64h4sq44wfz8vg 295938732632400000000000aepix
epixd add-genesis-account epix1jngqtljx0qg4cvc26lrgjzu73ejp3ch7aad62j 329880599466250000000000aepix
epixd add-genesis-account epix1sqd8pr8vfpagr46j8flps4az8rqdxrdwwntlmq 271164311619300000000000aepix
epixd add-genesis-account epix133geakf970n9ekfr8j6h68gaf08fj3srwl04lk 9571230649430000000000aepix
epixd add-genesis-account epix1l3lce3czu6y72562yj9kxqlgnn4lfpfpqwawy2 297290192890000000000aepix
epixd add-genesis-account epix1u3an8tv2n52xavn6ucdg0wa0qdxqlx8calzvhx 200240954887050000000000aepix
epixd add-genesis-account epix1lcjavr9tfxnsg0dx4xu9a8erdzdutj3z5daxr5 1213426312030350000000000aepix
epixd add-genesis-account epix1x0skmhahxz9js6l56v0u7zumhvx95ejcd606z5 35205869990000000000aepix
epixd add-genesis-account epix1532q0up9dch52la0d3k34ke2m2gcj5sr6m7sen 102071875950000000000aepix
epixd add-genesis-account epix1xfypuunnck4x7nzzyt6cvr0dq6gs0vcaqr9gky 266382700650000000000aepix
epixd add-genesis-account epix1g3l5tp723ueuy5zhv69ss6m9xs262x5475ly62 18372937671000000000000aepix
epixd add-genesis-account epix13e0uprdkpcx663n2a00r08nu6cdmr6xcn0lr6j 47658437689500000000000aepix
epixd add-genesis-account epix19hh7yda53gvvdprkqrsw0xu766ppspy9ml7ws5 264505573955700000000000aepix
epixd add-genesis-account epix1mrauk2206scmmuauwysn3emv9sm4zart75x86m 49505845041700000000000aepix
epixd add-genesis-account epix17cygjnrcjewvu5ecrr3m7gsq5qxle0lk0fzsjm 210187153823620000000000aepix
epixd add-genesis-account epix192p9m07aclj4qln38ym5rw5tw52z3lmc6e4cw6 127778364303330000000000aepix
epixd add-genesis-account epix1h6m4k69pla063xghz0jlkeshczp0mraygfcyv9 260148847543200000000000aepix
epixd add-genesis-account epix1g46np67sqsrxhssqdd55zuvcreyj0zq50ra5kf 264947841750140000000000aepix
epixd add-genesis-account epix17rqev8c7n5errck4z5rm40w03nu5y3s2p3zhga 6490047429790000000000aepix
epixd add-genesis-account epix1hhwnlwwd9kaychd2d0vz85etgqc3gj8gw5ds86 230491305264130000000000aepix
epixd add-genesis-account epix1x48cf94h3240gf6futgtqp50k7zczeswd8qsyu 57341162142790000000000aepix
epixd add-genesis-account epix1qpwrdm09tjveakega8rmluk2fs2f3awsdeyvve 54209302807120000000000aepix
epixd add-genesis-account epix1vl8auezt2w3k65py7062gccfpw20q2mw7ylqeg 15419523053620000000000aepix
epixd add-genesis-account epix1uw4gsx2yf2tps7aldl50h065vw2ppknfz7wlv2 33674019531790000000000aepix
epixd add-genesis-account epix1k7rhggdcsdk0l5r8vpp0kd958jrfla7ntqgg9t 15379519412700000000000aepix
epixd add-genesis-account epix17axwke8tsjs64akmnxj8hvdegn2v8e4904ge3p 46637727228520000000000aepix
epixd add-genesis-account epix10lsfrk8zqxfga0vy7lw7qm05v3sec4ue28y84q 102157785347580000000000aepix
epixd add-genesis-account epix1ctj9hwwsr5m2g2dqzhw0hnafqqsds4thxuqz93 513541579150290000000000aepix
epixd add-genesis-account epix1n7g5hqnxjtujpe3pkxfhp42qx8szgamwelzk9a 74862770642020000000000aepix

        # Update total supply with claim values
        total_supply=23689538000000000000000000
cat $HOME/.epixd/config/genesis.json | jq -r --arg total_supply "$total_supply" '.app_state["bank"]["supply"][0]["amount"]=$total_supply' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json

        # Configure distribution module with community pool
cat $HOME/.epixd/config/genesis.json | jq '.app_state["distribution"]["params"]["community_tax"]="0.000000000000000000"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["distribution"]["fee_pool"]["community_pool"] = [{"denom":"aepix","amount":"11844767999900000000000000"}]' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["bank"]["balances"] += [{"address":"epix1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8j52fwy","coins":[{"denom":"aepix","amount":"11844767999900000000000000"}]}]' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json

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


# Configure slashing parameters
cat $HOME/.epixd/config/genesis.json | jq '.app_state["slashing"]["params"]["signed_blocks_window"]="21600"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["slashing"]["params"]["min_signed_per_window"]="0.050000000000000000"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["slashing"]["params"]["downtime_jail_duration"]="60s"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["slashing"]["params"]["slash_fraction_double_sign"]="0.050000000000000000"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["slashing"]["params"]["slash_fraction_downtime"]="0.010000000000000000"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json


# Configure onboarding module
cat $HOME/.epixd/config/genesis.json | jq '.app_state["onboarding"]["params"]["enable_onboarding"]=true' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["onboarding"]["params"]["whitelisted_channels"]=["channel-0"]' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["onboarding"]["params"]["auto_swap_threshold"]="10000000000000000"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json

# Configure IBC parameters
cat $HOME/.epixd/config/genesis.json | jq '.app_state["ibc"]["connection_genesis"]["params"]["max_expected_time_per_block"]="30000000000"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["transfer"]["params"]["send_enabled"]=true' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["transfer"]["params"]["receive_enabled"]=true' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["transfer"]["port_id"]="transfer"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json

# Configure fee market parameters
cat $HOME/.epixd/config/genesis.json | jq '.app_state["feemarket"]["params"]["min_gas_price"]="0.0000001"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json

# Configure auth params
cat $HOME/.epixd/config/genesis.json | jq '.app_state["auth"]["params"]["max_memo_characters"]="256"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["auth"]["params"]["tx_sig_limit"]="7"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["auth"]["params"]["tx_size_cost_per_byte"]="10"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["auth"]["params"]["sig_verify_cost_ed25519"]="590"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["auth"]["params"]["sig_verify_cost_secp256k1"]="1000"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["feemarket"]["params"]["base_fee"]="100000000"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["feemarket"]["params"]["enable_height"]="0"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["feemarket"]["params"]["no_base_fee"]=false' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["feemarket"]["params"]["elasticity_multiplier"]="2"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json

# Configure EVM fees
cat $HOME/.epixd/config/genesis.json | jq '.app_state["evm"]["params"]["evm_denom"]="aepix"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["evm"]["params"]["chain_config"]["london_block"]="0"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["evm"]["params"]["chain_config"]["arrow_glacier_block"]="0"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["evm"]["params"]["chain_config"]["gray_glacier_block"]="0"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["evm"]["params"]["chain_config"]["merge_netsplit_block"]="0"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json

# Configure IBC parameters
cat $HOME/.epixd/config/genesis.json | jq '.app_state["ibc"]["connection_genesis"]["params"]["max_expected_time_per_block"]="30000000000"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["transfer"]["params"]["send_enabled"]=true' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["transfer"]["params"]["receive_enabled"]=true' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json
cat $HOME/.epixd/config/genesis.json | jq '.app_state["transfer"]["port_id"]="transfer"' > $HOME/.epixd/config/tmp_genesis.json && mv $HOME/.epixd/config/tmp_genesis.json $HOME/.epixd/config/genesis.json

        # Sign genesis transaction
        epixd gentx $KEY 1000000000000000000aepix --keyring-backend $KEYRING --chain-id $CHAINID --fees 1aepix --gas 200000
        
        # Collect genesis tx
        epixd collect-gentxs
        
        # Run this to ensure everything worked and that the genesis file is setup correctly
        epixd validate-genesis
        
        if [[ $1 == "pending" ]]; then
          echo "pending mode is on, please wait for the first block committed."
        fi
        
        # Start the node (remove the --pruning=nothing flag if historical queries are not needed)
        epixd start --pruning=nothing $TRACE --log_level $LOGLEVEL --minimum-gas-prices=0.0000001aepix --json-rpc.api eth,txpool,personal,net,debug,web3 --rpc.laddr "tcp://0.0.0.0:26657" --api.enable --chain-id $CHAINID
        