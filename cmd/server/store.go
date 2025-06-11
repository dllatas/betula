package main

import (
	"fmt"
	"log/slog"

	"github.com/dllatas/betula/lib"
)

func setupStore(dataPath string) (*lib.Store, error) {
	var store *lib.Store

	fileExists, err := lib.ExistsFile(dataPath)
	if err != nil {
		return nil, fmt.Errorf("betula: failed to check data path existence %s", err.Error())
	}

	slog.Info("setup store", "datapath", dataPath, "data file exists", fileExists)

	if fileExists {
		store, err = lib.LoadStoreFromDataFileIfExists(dataPath)
		if err != nil {
			return nil, fmt.Errorf("betula: failed to load data from existing data filepath %s", err.Error())
		}
	} else {
		err = lib.CreateFile(dataPath)
		if err != nil {
			return nil, fmt.Errorf("betula: failed to create data path %s", err.Error())
		}

		store = lib.NewInMemoryStore()
	}

	store.DataPath = dataPath

	wal, err := lib.NewWAL(dataPath)
	if err != nil {
		return nil, fmt.Errorf("betula: wal file failed to init %s", err.Error())
	}

	walEntries, err := wal.Replay()
	if err != nil {
		return nil, fmt.Errorf("betula: wal replay failed %s", err.Error())
	}

	err = store.LoadWALEntries(walEntries)
	if err != nil {
		return nil, fmt.Errorf("betula: load wal entries failed %s", err.Error())
	}

	err = store.SaveToFile()
	if err != nil {
		return nil, fmt.Errorf("betula: save checkpoint %s", err.Error())
	}

	err = wal.Truncate()
	if err != nil {
		return nil, fmt.Errorf("betula: truncate wal failed %s", err.Error())
	}

	store.WAL = wal

	return store, nil
}
