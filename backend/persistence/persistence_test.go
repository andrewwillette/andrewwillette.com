package persistence

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const testDatabaseFile = "testDatabase.db"

type PersistenceTestSuite struct {
	suite.Suite
}

func TestPersistenceSuite(t *testing.T) {
	suite.Run(t, new(PersistenceTestSuite))
}

func (suite *PersistenceTestSuite) SetupTest() {
	deleteTestDatabase()
}

func (suite *PersistenceTestSuite) TearDownSuite() {
	deleteTestDatabase()
}

func deleteTestDatabase() {
	_ = os.Remove(testDatabaseFile)
}

func (suite *PersistenceTestSuite) TestExecuteSqlScript() {
	if err := createDatabase(testDatabaseFile); err != nil {
		suite.T().Fail()
	}
	_ = runSqlScript(testDatabaseFile, "./sql/createUserTable.sql")
	tables, err := getAllTables(testDatabaseFile)
	if err != nil {
		suite.T().Log("failed to get all tables")
		suite.T().Fail()
	}
	suite.T().Log(len(tables))
	assert.ElementsMatch(suite.T(), tables, [2]string{"userCredentials", "sqlite_sequence"})
	_ = runSqlScript(testDatabaseFile, "./sql/createSoundcloudTable.sql")
	tables, err = getAllTables(testDatabaseFile)
	suite.T().Log(len(tables))
	if err != nil {
		suite.T().Log("failed to get all tables")
		suite.T().Fail()
	}
	assert.ElementsMatch(suite.T(), tables, [3]string{"userCredentials", "sqlite_sequence", "soundcloudUrl"})
}

func (suite *PersistenceTestSuite) TestModifySoundcloudTableUiOrder() {
	if err := createDatabase(testDatabaseFile); err != nil {
		suite.T().Fail()
	}
	_ = runSqlScript(testDatabaseFile, "./sql/createSoundcloudTable.sql")
	columnNames, _ := getTableColumnNames(testDatabaseFile, soundcloudTable)
	assert.ElementsMatch(suite.T(), columnNames, [2]string{"id", "url"})
	_ = runSqlScript(testDatabaseFile, "./sql/alterSoundcloudTableUiOrder.sql")
	columnNames, _ = getTableColumnNames(testDatabaseFile, soundcloudTable)
	assert.ElementsMatch(suite.T(), columnNames, [3]string{"id", "url", "uiOrder"})
}

func (suite *PersistenceTestSuite) TestCreateDatabase() {
	err := createDatabase(testDatabaseFile)
	if err != nil {
		suite.T().Log("failed to create database")
		suite.T().Fail()
	}
	_, err = sql.Open("sqlite3", fmt.Sprintf("./%s", testDatabaseFile))
	if err != nil {
		suite.T().Log("failed to open sqlite database")
		suite.T().Fail()
	}
	assert.FileExists(suite.T(), testDatabaseFile)
}

// FuzzCreateDatabase test CreateDatabase fuzz handling
// I don't think this is a good test. There are occurrences where
// trying to create afile from a string should throw an error.
// Fuzzing over strings will eventually throw an error, it's wrong to
// assume otherwise. Makes we wonder about fuzzing in general. Aren't
// there always workflows where your function should throw an error?
func FuzzCreateDatabase(f *testing.F) {
	f.Add("swag.sqlite")
	f.Fuzz(func(t *testing.T, filename string) {
		err := createDbString(filename)
		require.NoError(t, err)
	})
}
