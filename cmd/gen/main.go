package main

import (
	"context"
	"encoding/json"
	"fmt"
	log "github.com/xlab/suplog"
	"io/ioutil"
	"os"
	"time"
)



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
	// fill metas for each
	addressMap := buildAddressMap()
	for token := range tokenMetaMap {
		tokenAddressHex := tokenMetaMap[token].Address
		if tokenAddressHex == "" {
			continue
		}
		mainnetAddressHex := addressMap[tokenAddressHex]

		if mainnetAddressHex != "" {
			log.Infof("fetching token [%s] from Alchemy using [%s] instead of [%s]\n", token, mainnetAddressHex, tokenAddressHex)
			tokenAddressHex = mainnetAddressHex
		}
		metadata := getTokenMetaByAddress(ctx, tokenAddressHex)
		if metadata == nil {
			log.Panicf("token [%s] metadata is empty, address: [%s]\n", token, tokenAddressHex)
		}
		log.Infof("got metadata for token [%s]\n", token)
		tokenMetaMap[token].Meta = metadata
	}
	log.Infof("finished fetching tokens' metadata\n")

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
