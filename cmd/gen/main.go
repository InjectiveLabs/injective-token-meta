package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bugsnag/bugsnag-go/errors"
	log "github.com/xlab/suplog"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

func init() {
	readEnv()
	alchemyAPIKey = os.Getenv(alchemyAPIKeyEnvVar)
	coinmarketcapAPIKey = os.Getenv(coinmarketcapAPIKeyEnvVar)
}

func main() {
	ctx := context.Background()

	tokenMetaFile, err := os.OpenFile(tokenMetaFilePath, os.O_RDWR, os.ModePerm)
	orPanicf(err, "failed to open support token list file\n")
	defer func() {
		orPanicf(tokenMetaFile.Close(), "failed to close support token list file\n")
	}()
	fileContent, err := ioutil.ReadAll(tokenMetaFile)
	orPanicf(err, "failed to read support token list file\n")

	// copy for back up
	err = ioutil.WriteFile(fmt.Sprintf(tokenMetaBackUpFilePathPattern, time.Now().UTC().Format(time.RFC3339)),
		fileContent, 0644)
	orPanicf(err, "failed to write back up file\n")

	tokenMetaMap := TokenMetaMap{}
	err = json.Unmarshal(fileContent, &tokenMetaMap)
	orPanicf(err, "failed to json unmarshal token meta file content\n")

	log.Infof("got token meta map, [%d] tokens' metadata need to be filled\n", len(tokenMetaMap))
	log.Infof(logDivider)
	tokenMetaMap.tidy()

	// fill metas for each
	for token := range tokenMetaMap {
		err := fillTokenMeta(ctx, token, tokenMetaMap[token])
		orPanicf(err, "failed to fill token meta for [%s]\n", token)
		log.Infof("successfully filled token meta for [%s]\n", token)
	}
	log.Infof("finished fetching tokens' metadata\n")
	log.Infof(logDivider)
	for k, v := range tokenMetaCustomizerMap {
		v(tokenMetaMap[k])
	}
	log.Infof("finished customized metadata for [%d] tokens\n", len(tokenMetaMap))
	log.Infof(logDivider)

	tokenMetaMap.check()

	// write token metadata map to file
	newFileContent, err := json.MarshalIndent(&tokenMetaMap, "", "  ")
	orPanicf(err, "failed to json marshal new token meta map\n")

	orPanicf(tokenMetaFile.Truncate(0), "failed to truncate token meta file\n")
	_, err = tokenMetaFile.Seek(0, 0)
	orPanicf(err, "failed to reset the offset to write the token meta file\n")

	_, err = tokenMetaFile.Write(newFileContent)
	orPanicf(err, "failed to write new file content\n")
	orPanicf(tokenMetaFile.Sync(), "failed to sync new file\n")

	log.Infof("successfully gen token meta file\n")
}

func orPanicf(err error, format string, args ...interface{}) {
	if err != nil {
		log.WithError(err).Panicf(format, args...)
	}
}

func fillTokenMeta(ctx context.Context, symbol string, t *Token) error {
	if t == nil {
		return errors.Errorf("empty token\n")
	}

	if t.CoingeckoID == "" {
		return errors.Errorf("empty coingecko id, might cause an error when query token's price\n")
	} else {
		coin := GetCoingeckoTokenDetail(t.CoingeckoID)
		if strings.ToLower(coin.AssetPlatformId) != ethereum {
			log.Warningf("token [%s] platform [%s] is not %s\n", symbol, coin.AssetPlatformId, ethereum)
		}
		if strings.ToLower(coin.Platforms[ethereum]) != strings.ToLower(t.Address) {
			log.Warningf("token [%s] address [%s] is not same as in coingecko resp [%s], platforms: [%+v]\n",
				symbol, t.Address, coin.Platforms[ethereum], coin.Platforms)
		}
		// address is valid
	}
	if t.Address == "" {
		addr := GetEthereumAddressBySymbol(symbol)
		if addr == "" {
			log.Warningf("cannot solve ethereum address from symbol [%s], better to cover this in customizers\n", symbol)
			return nil
		}
		t.Address = strings.ToLower(addr)
	}

	metadata := getTokenMetaByAddress(ctx, t.Address)
	if metadata == nil {
		log.Panicf("token metadata is empty, address: [%s]\n", t.Address)
	}
	t.Meta = metadata

	if t.Meta.Logo == "" {
		t.Meta.Logo = GetLogoBySymbol(symbol)
	}
	return nil
}

func readEnv() {
	if envdata, _ := ioutil.ReadFile(".env"); len(envdata) > 0 {
		s := bufio.NewScanner(bytes.NewReader(envdata))
		for s.Scan() {
			txt := s.Text()
			valIdx := strings.IndexByte(txt, '=')
			if valIdx < 0 {
				continue
			}

			strValue := strings.Trim(txt[valIdx+1:], `"`)
			if err := os.Setenv(txt[:valIdx], strValue); err != nil {
				log.WithField("name", txt[:valIdx]).WithError(err).Warningln("failed to override ENV variable")
			}
		}
	}
}
