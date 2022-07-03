package config

import (
	"errors"
	"os"
	"path/filepath"
)

const (
	Port = 9099
)

func GetDatabaseFile() string {
	const dockerDbDir = "/awillettebackend/db"
	const dbFile = "sqlite-database.db"
	_, err := os.Stat(dockerDbDir)
	if errors.Is(err, os.ErrNotExist) {
		return dbFile
	}
	return filepath.Join(dockerDbDir, dbFile)
}
