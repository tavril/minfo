package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

var errEmptyCache = errors.New("cache file is empty")

func readCacheFile(cacheFilePath string) (err error) {
	var fileInfo os.FileInfo
	if fileInfo, err = os.Stat(cacheFilePath); err != nil {
		fmt.Printf("Error reading cache file: %v\n", err)
		return
	}
	if fileInfo.Size() == 0 {
		return errEmptyCache
	}
	var cacheData []byte
	if cacheData, err = os.ReadFile(cacheFilePath); err != nil {
		return
	}
	err = json.Unmarshal(cacheData, &hostInfo)
	return
}

func writeCacheFile(cacheFilePath string) (err error) {
	var jsonData []byte
	if jsonData, err = json.MarshalIndent(hostInfo.cachedInfo, "", "  "); err != nil {
		return
	}
	err = os.WriteFile(cacheFilePath, jsonData, 0644)
	return
}
