package lib

import (
	"fmt"
	"os"
	"path/filepath"
)

func ExistsFile(path string) (bool, error) {
	dir := filepath.Dir(path)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return false, nil
	}

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	if info.IsDir() {
		return false, fmt.Errorf("exists file: path %s is a dir", path)
	}

	return true, nil
}

func CreateFile(path string) error {
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return err
	}

	return nil
}
