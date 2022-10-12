package config

import (
	"errors"
	"os"
	"path/filepath"
)

const (
	dockerDbDir = "/awillettebackend/db"
	localDbFile = "sqlite-database.db"
)

func GetDatabaseFile() string {
	_, err := os.Stat(dockerDbDir)
	if errors.Is(err, os.ErrNotExist) {
		return localDbFile
	}
	return filepath.Join(dockerDbDir, localDbFile)
}
