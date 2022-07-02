package persistence

import (
	"database/sql"
	"fmt"
	"github.com/andrewwillette/willette_api/logging"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"os"
)

func runSqlScript(databaseFile, sqlScriptFilePath string) error {
	sqldb, err := sql.Open("sqlite3", fmt.Sprintf("./%s", databaseFile))
	if err != nil {
		logging.GlobalLogger.Err(err).Msg("failed to open sqlite file in runSqlScript")
		return err
	}
	stringEx, err := ioutil.ReadFile(sqlScriptFilePath)
	if err != nil {
		logging.GlobalLogger.Err(err).Msg("failed to read sqlScriptFilePath to run")
		return err
	}
	sqlStatement := string(stringEx)
	preparedStatement, err := sqldb.Prepare(sqlStatement)
	if err != nil {
		logging.GlobalLogger.Err(err).Msg("failed to prepare sql script")
		return err
	}
	_, err = preparedStatement.Exec()
	if err != nil {
		logging.GlobalLogger.Err(err).Msg("failed to execute sql statement")
		return err
	}
	return nil
}

func createDatabase(databaseFile string) error {
	file, err := os.Create(databaseFile)
	if err != nil {
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	sqliteDatabase, err := sql.Open("sqlite3", fmt.Sprintf("./%s", databaseFile))
	if err != nil {
		return err
	}
	if err = sqliteDatabase.Close(); err != nil {
		return err
	}
	return nil
}

func getQueryResponseAsMap(sqlQuery *sql.Stmt) ([]map[string]interface{}, error) {
	rows, err := sqlQuery.Query()
	if err != nil {
		logging.GlobalLogger.Err(err).Msg("Error executing getQueryResponseAsMap sql query.")
		return nil, err
	}
	// very important, returns the column names
	columnNames, err := rows.Columns()
	if err != nil {
		logging.GlobalLogger.Err(err).Msg("Error getting column names.")
		return nil, err
	}
	allRows := make([]map[string]interface{}, 0)
	columnValues := make([]interface{}, len(columnNames))
	defer rows.Close()
	for rows.Next() {
		rowAsMap := make(map[string]interface{}, len(columnNames))
		for i := range columnValues {
			columnValues[i] = new(interface{})
		}
		if err := rows.Scan(columnValues...); err != nil {
			logging.GlobalLogger.Err(err).Msg("Error getting column values")
			return nil, err
		}
		for i, col := range columnNames {
			rowAsMap[col] = *columnValues[i].(*interface{})
		}
		allRows = append(allRows, rowAsMap) // unnecessary but keeping for future reference
	}
	return allRows, nil
}

func getTableColumnNames(databaseFile, tableName string) ([]string, error) {
	sqliteDatabase, _ := sql.Open("sqlite3", fmt.Sprintf("./%s", databaseFile))
	getAllProperties := fmt.Sprintf("pragma table_info(%s)", tableName)
	addSoundcloudPreparedStatement, err := sqliteDatabase.Prepare(getAllProperties)
	if err != nil {
		logging.GlobalLogger.Err(err).Msg("Error preparing getAllColumnNames sql.")
		return nil, err
	}
	values, err := getQueryResponseAsMap(addSoundcloudPreparedStatement)
	var columnNames []string
	for _, v := range values {
		columnNames = append(columnNames, fmt.Sprint(v["name"]))
	}
	return columnNames, nil
}

func getAllTables(databaseFile string) ([]string, error) {
	sqliteDatabase, _ := sql.Open("sqlite3", fmt.Sprintf("./%s", databaseFile))
	rows, err := sqliteDatabase.Query("SELECT name FROM sqlite_master WHERE type='table';")
	if err != nil {
		return nil, err
	}
	var table string
	var tables []string
	for rows.Next() {
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func alterSoundcloudTableUiOrderAddition(sqlDbFile string) error {
	if err := runSqlScript(sqlDbFile, "./persistence/sql/alterSoundcloudTableUiOrder.sql"); err != nil {
		logging.GlobalLogger.Err(err).Msg("Failed to run run script: alterSoundcloudTableUiOrder.sql")
		return err
	}
	return nil
}

// InitDatabaseIdempotent - Creates database if it doesn't currently exist.
// Database creation includes creating the user and soundcloud table.
func InitDatabaseIdempotent(sqliteFile string) {
	if _, err := os.Stat(sqliteFile); os.IsNotExist(err) {
		if err = createDatabase(sqliteFile); err != nil {
			logging.GlobalLogger.Fatal().Msg("failed to create database")
		}
		userService := &UserService{SqliteDbFile: sqliteFile}
		userService.createUserTable()
		soundcloudUrlService := &SoundcloudUrlService{SqliteFile: sqliteFile}
		soundcloudUrlService.createSoundcloudUrlTable()
		//_ = alterSoundcloudTableUiOrderAddition(sqliteFile)
	}
}
