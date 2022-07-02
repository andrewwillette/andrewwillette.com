package persistence

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type UsersTestSuite struct {
	suite.Suite
}

func TestUsersSuite(t *testing.T) {
	suite.Run(t, new(UsersTestSuite))
}

func (suite *UsersTestSuite) SetupTest() {
	deleteTestDatabase()
}

func (suite *UsersTestSuite) TearDownSuite() {
	deleteTestDatabase()
}

func (suite *UsersTestSuite) TestCreateUserTable() {
	userService := &UserService{SqliteDbFile: testDatabaseFile}
	userService.createUserTable()
	tables, err := getAllTables(testDatabaseFile)
	if err != nil {
		suite.T().Fail()
	}
	assert.Equal(suite.T(), tables[0], "userCredentials")
}

func (suite *UsersTestSuite) TestCreateUser_Valid() {
	userService := &UserService{SqliteDbFile: testDatabaseFile}
	userService.createUserTable()
	username := "usernameOne"
	password := "passwordOne"
	err := userService.addUser(username, password)
	if err != nil {
		suite.T().Logf("failed to add user")
		suite.T().Fail()
	}
	users, err := userService.getAllUsers()
	assert.Equal(suite.T(), users[0].Username, username)
	assert.Equal(suite.T(), users[0].Password, password)
}

func (suite *UsersTestSuite) TestUpdateUserBearerToken_Valid() {
	userService := &UserService{SqliteDbFile: testDatabaseFile}
	userService.createUserTable()
	username := "usernameOne"
	password := "passwordOne"
	err := userService.addUser(username, password)
	if err != nil {
		suite.T().Logf("failed to add user")
		suite.T().Fail()
	}
	bearerToken := "bearerTokenOne"
	userService.updateUserBearerToken(username, password, bearerToken)
	user, err := userService.getUser(username, password)
	if err != nil {
		suite.T().Log(err)
		suite.T().Fail()
	}
	assert.Equal(suite.T(), user.Username, username)
	assert.Equal(suite.T(), user.Password, password)
	assert.Equal(suite.T(), user.BearerToken, bearerToken)

	userExists := userService.userExists(&User{Username: username, Password: password})
	if err != nil {
		suite.T().Log(err)
		suite.T().Fail()
	}
	assert.True(suite.T(), userExists)

	bearerTokenExists := userService.IsAuthorized(bearerToken)
	assert.True(suite.T(), bearerTokenExists)
}
