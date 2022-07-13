package config

import (
	"errors"
	"os"
	"path/filepath"
)

const (
	Port        = 9099
	dockerDbDir = "/awillettebackend/db"
	dbFile      = "sqlite-database.db"
)

func GetDatabaseFile() string {
	_, err := os.Stat(dockerDbDir)
	if errors.Is(err, os.ErrNotExist) {
		return dbFile
	}
	return filepath.Join(dockerDbDir, dbFile)
}
