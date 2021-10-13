# injective-token-meta

Token metadata of Injective Exchange

## Usage
The static token meta file is in `meta/token_meta.json`.<br>
Different repo can import this repo and use the json file in its own way.<br>
For Go repo, please import `utils-go` and use the util functions in it. 

## Maintain
To add a new token:
1. add a new `token_name: {address:token_address, coinGeckoId:token_coingecko_id}` kv pair in `token_meta.json`
2. run `make gen`
