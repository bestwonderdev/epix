<!--
order: 4
-->

# Parameters

The coinswap module contains the following parameters:

| Key                    | Type         | Default value                                                                                                                                                                                                                                                                                                              |
|:-----------------------|:-------------|:---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Fee                    | string (dec) | "0.0"                                                                                                                                                                                                                                                                                                                      |
| PoolCreationFee        | sdk.Coin     | "0aepix"                                                                                                                                                                                                                                                                                                                  |
| TaxRate                | string (dec) | "0.0"                                                                                                                                                                                                                                                                                                                      |
| MaxStandardCoinPerPool | string (int) | "10000000000000000000000"                                                                                                                                                                                                                                                                                                  |
| MaxSwapAmount          | sdk.Coins    | [{"denom":"ibc/8C2E3CD84DB23CCFB34A14E4EA933B6C38F9EA3FD7187E405C889C66C4347D31","amount":"10000000"},{"denom":"ibc/1A9E6C67B20E24F2F2E551D1E1B6E27F0D171B92B4215F0338F0F0D931A98264","amount":"10000000"},{"denom":"ibc/3F45E3F66EF983C7A131C4187BF49E5D98598DBF3E07C14F8F9F573A4E975834","amount":"100000000000000000"}] |

### Fee

Swap fee rate for swap. In this version, swap fees aren't paid upon swap orders directly. Instead, pool just adjust pool's quoting prices to reflect the swap fees.

### PoolCreationFee

Fee paid for to create a pool. This fee prevents spamming and is collected in the fee collector.

### TaxRate

Community tax rate for pool creation fee. This tax is collected in the fee collector.

### MaxStandardCoinPerPool

Maximum amount of standard coin per pool. This parameter is used to prevent pool from being too large.

### MaxSwapAmount

Maximum amount of swap amount. This parameter is used to prevent swap from being too large. It is also used as whitelist for pool creation.
