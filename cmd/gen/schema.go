package main

type Token struct {
	Address     string `json:"address"`
	CoingeckoID string `json:"coinGeckoId"`
	Meta        *Meta  `json:"meta"`
}

// Meta this struct is the same as the metadata in the resp of Alchemy
type Meta struct {
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
	Logo     string `json:"logo"`
}

type TokenMetaMap map[string]*Token
