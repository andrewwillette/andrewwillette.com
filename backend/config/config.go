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
	const dockerDatabase = "/goApp/db"
	_, err := os.Stat(dockerDatabase)
	if errors.Is(err, os.ErrNotExist) {
		return "sqlite-database.db"
	}
	return fmt.Sprintf("%s/%s", dockerDatabase, "sqlite-database.db")
}

func GetCorsWhiteList() []string {
	return []string{"http://localhost:3000", "http://andrewwillette.com"}
}
