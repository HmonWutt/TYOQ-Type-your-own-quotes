package db

import (
	_ "embed"
	"os"
	"path/filepath"
)

//go:embed seed.db
var seedDB []byte

func DefaultDBPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "tyoq", "seed.db"), nil
}

func EnsureDB() (string, error) {
	path, err := DefaultDBPath()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(path); err == nil {
		return path, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, seedDB, 0o644); err != nil {
		return "", err
	}
	return path, nil
}
