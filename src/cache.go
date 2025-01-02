package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

var errEmptyCache = errors.New("cache file is empty")

func readCacheFile(cacheFilePath string) (err error) {
	var fileInfo os.FileInfo
	if fileInfo, err = os.Stat(cacheFilePath); err != nil {
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

func writeCacheFile(cacheFilePath string, hostInfo *info) (err error) {
	dirPath := filepath.Dir(cacheFilePath)
	if err = ensureDirExists(dirPath); err != nil {
		return
	}
	var jsonData []byte
	if jsonData, err = json.MarshalIndent((*hostInfo).cachedInfo, "", "  "); err != nil {
		return
	}
	err = os.WriteFile(cacheFilePath, jsonData, 0644)
	return
}

func populateCache(cacheFilePath string) (err error) {
	spDataTypes := map[string]bool{}
	var items []string
	for k, v := range availableItems {
		if v.SPDataType != nil && v.IsCached {
			if _, ok := spDataTypes[*v.SPDataType]; !ok {
				spDataTypes[*v.SPDataType] = true
			}
			items = append(items, k)
		}
	}

	spErr := fetchSystemProfiler(&hostInfo, items, spDataTypes, false)
	if spErr != nil {
		return spErr
	}

	if err := writeCacheFile(cacheFilePath, &hostInfo); err != nil {
		return err
	}
	return nil
}
