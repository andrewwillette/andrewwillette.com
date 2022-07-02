package config

import (
	"errors"
	"fmt"
	"os"
)

const (
	/*SqliteFile = "sqlite-database.db"*/
	Port = 9099
)

func GetDatabaseFile() string {
	const dockerDbDir = "/awillettebackend/db"
	_, err := os.Stat(dockerDbDir)
	if errors.Is(err, os.ErrNotExist) {
		return "sqlite-database.db"
	}
	return fmt.Sprintf("%s/%s", dockerDbDir, "sqlite-database.db")
}
