package main

import (
	"fmt"
	"log/slog"

	"github.com/dllatas/betula/lib"
)

func setupStore(dataPath string) (*lib.Store, error) {
	slog.Info("setup store", "datapath", dataPath)

	// 1. Check if checkpoint file exists
	checkpointExists, err := lib.ExistsFile(dataPath)
	if err != nil {
		return nil, fmt.Errorf("betula: failed to check data path existence %s", err.Error())
	}

	// 2. Init store
	var store *lib.Store
	if checkpointExists {
		slog.Info("checkpoint exists, loading from file", "path", dataPath)
		store, err = lib.LoadStoreFromDataFileIfExists(dataPath)
		if err != nil {
			return nil, fmt.Errorf("betula: failed to load checkpoint (store) %s", err.Error())
		}
	} else {
		slog.Warn("no checkpoint found, creating new store and file", "path", dataPath)
		err = lib.CreateFile(dataPath)
		if err != nil {
			return nil, fmt.Errorf("betula: failed to create data path %s", err.Error())
		}

		store = lib.NewInMemoryStore()
	}

	store.DataPath = dataPath

	// 3. Init WAL
	wal, err := lib.NewWAL(dataPath)
	if err != nil {
		return nil, fmt.Errorf("betula: wal file failed to init %s", err.Error())
	}
	store.WAL = wal

	// 4. WAL replay
	walEntries, err := wal.Replay()
	if err != nil {
		return nil, fmt.Errorf("betula: wal replay failed %s", err.Error())
	}

	if len(walEntries) > 0 {
		slog.Info("replaying WAL entries", "count", len(walEntries))
		if err := store.LoadWALEntries(walEntries); err != nil {
			return nil, fmt.Errorf("betula: failed to load WAL entries: %w", err)
		}
	} else {
		slog.Info("betula: no WAL entries to replay")
	}

	// 5. Persist store & truncate WAL
	err = store.SaveToFile()
	if err != nil {
		return nil, fmt.Errorf("betula: save checkpoint %s", err.Error())
	}

	err = wal.Truncate()
	if err != nil {
		return nil, fmt.Errorf("betula: truncate WAL failed %s", err.Error())
	}

	slog.Info("store setup completed")

	return store, nil
}
