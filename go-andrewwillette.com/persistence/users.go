package persistence

import (
	"database/sql"
	"fmt"
	"github.com/andrewwillette/willette_api/logging"
)

const userTable = "userCredentials"

type UserService struct {
	SqliteDbFile string
}

type User struct {
	Username    string
	Password    string
	BearerToken string
}

func (u *UserService) createUserTable() {
	db, err := sql.Open("sqlite3", u.SqliteDbFile)
	if err != nil {
		logging.GlobalLogger.Err(err).Msg("Error opening database in create user table sql")
	}
	defer db.Close()
	createUserTableSQL := fmt.Sprintf("CREATE TABLE %s ("+
		"\"username\" TEXT NOT NULL, "+
		"\"password\" TEXT NOT NULL, "+
		"\"bearerToken\" BLOB"+
		")", userTable)
	statement, err := db.Prepare(createUserTableSQL)
	if err != nil {
		logging.GlobalLogger.Err(err).Msg("Error preparing create user table sql")
	}
	_, err = statement.Exec()
	if err != nil {
		logging.GlobalLogger.Err(err).Msg("Error executing create user table sql")
	}
}

func (u *UserService) addUser(username, password string) error {
	db, err := sql.Open("sqlite3", u.SqliteDbFile)
	if err != nil {
		logging.GlobalLogger.Err(err).Msg("Error opening database in addUser")
		return err
	}
	addUserSqlStatement := fmt.Sprintf("INSERT INTO %s(username, password) "+
		"VALUES('%s', '%s');", userTable, username, password)
	addUserStatement, err := db.Prepare(addUserSqlStatement)
	if err != nil {
		logging.GlobalLogger.Err(err).Msg("Error preparing add user sql")
		return err
	}
	_, err = addUserStatement.Exec()
	if err != nil {
		logging.GlobalLogger.Err(err).Msg("Error executing sql in addUser")
		return err
	}
	return nil
}

// UpdateUserBearerToken Adds the provided bearerToken to the username/password
func (u *UserService) updateUserBearerToken(username, password, bearerToken string) {
	db, err := sql.Open("sqlite3", u.SqliteDbFile)
	if err != nil {
		logging.GlobalLogger.Err(err).Msg("Error opening database in update bearer token")
		return
	}
	defer db.Close()
	addUserWithSessionKey := fmt.Sprintf("UPDATE %s SET bearerToken = '%s' WHERE username = '%s'",
		userTable, bearerToken, username)
	addUserWithSessionKeyStatement, err := db.Prepare(addUserWithSessionKey)
	if err != nil {
		logging.GlobalLogger.Err(err).Msg("Error preparing update user bearer token sql")
		return
	}
	_, err = addUserWithSessionKeyStatement.Exec()
	if err != nil {
		logging.GlobalLogger.Err(err).Msg("Error executing update user bearer token sql")
		return
	}
}

func (u *UserService) getAllUsers() ([]User, error) {
	db, err := sql.Open("sqlite3", u.SqliteDbFile)
	if err != nil {
		logging.GlobalLogger.Err(err).Msg("Error opening database in getAllUsers")
		return nil, err
	}
	selectAllUsers := fmt.Sprintf("SELECT * FROM %s", userTable)
	selectAllUsersPrepared, err := db.Prepare(selectAllUsers)
	if err != nil {
		logging.GlobalLogger.Err(err).Msg("Error preparing select all from users sql query")
		return nil, err
	}
	rows, err := selectAllUsersPrepared.Query()
	if err != nil {
		logging.GlobalLogger.Err(err).Msg("Error executing select all users sql query")
		return nil, err
	}
	defer rows.Close()
	var username, password, bearerToken sql.NullString
	var users []User
	for rows.Next() {
		err = rows.Scan(&username, &password, &bearerToken)
		if err != nil {
			logging.GlobalLogger.Err(err).Msg("Error scanning sql rows in getAllUsers")
			continue
		}
		user := User{Username: username.String, Password: password.String}
		users = append(users, user)
	}
	return users, err
}

func (u *UserService) getUser(username, password string) (User, error) {
	db, err := sql.Open("sqlite3", u.SqliteDbFile)
	if err != nil {
		logging.GlobalLogger.Err(err).Msg("Error opening database in get user")
		return User{}, err
	}
	defer db.Close()
	getUserStatement := fmt.Sprintf(`SELECT * FROM %s WHERE username = "%s" AND password = "%s" LIMIT 1`, userTable, username, password)
	preparedStatement, err := db.Prepare(getUserStatement)
	if err != nil {
		logging.GlobalLogger.Err(err).Msg("Error preparing getUser by username/password sql statement")
		return User{}, err
	}
	row := preparedStatement.QueryRow()
	user := User{}
	err = row.Scan(&user.Username, &user.Password, &user.BearerToken)
	if err != nil {
		logging.GlobalLogger.Err(err).Msg("Error scanning get user by username/password sql row")
		return User{}, err
	}
	return user, nil
}

// userExists Checks database if username, password exists
func (u *UserService) userExists(user *User) bool {
	db, err := sql.Open("sqlite3", u.SqliteDbFile)
	if err != nil {
		logging.GlobalLogger.Err(err).Msg("Error opening database")
		return false
	}
	defer db.Close()
	userExistsStatement := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE username = "%s" AND password = "%s")`, userTable, user.Username, user.Password)
	preparedStatement, err := db.Prepare(userExistsStatement)
	if err != nil {
		logging.GlobalLogger.Err(err).Msg("Error preparing user exists sql query")
		return false
	}
	rows, err := preparedStatement.Query()
	if err != nil {
		logging.GlobalLogger.Err(err).Msg("Error executing select user exists sql query")
		return false
	}
	defer rows.Close()
	var success string
	for rows.Next() {
		err := rows.Scan(&success)
		if err != nil {
			logging.GlobalLogger.Err(err).Msg("Error scanning rows of user exists sql query")
			return false
		}
		break
	}
	if success == "1" {
		return true
	} else {
		return false
	}
}

func (u *UserService) Login(username, password string) (success bool, authToken string) {
	user := User{Username: username, Password: password}
	userExists := u.userExists(&user)
	if userExists {
		authToken := newSHA1Hash()
		u.updateUserBearerToken(user.Username, user.Password, authToken)
		return true, authToken
	} else {
		return false, ""
	}
}

func (u *UserService) IsAuthorized(authToken string) bool {
	db, err := sql.Open("sqlite3", u.SqliteDbFile)
	if err != nil {
		logging.GlobalLogger.Err(err).Msg("Error opening database")
		return false
	}
	defer db.Close()
	bearerTokenExists := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE bearerToken = "%s")`, userTable, authToken)
	preparedStatement, err := db.Prepare(bearerTokenExists)
	if err != nil {
		logging.GlobalLogger.Err(err).Msg("Error preparing bearerTokenExists sql statement")
		return false
	}
	rows, err := preparedStatement.Query()
	if err != nil {
		logging.GlobalLogger.Err(err).Msg("Error executing bearerTokenExists sql query")
		return false
	}
	defer rows.Close()
	var success string
	for rows.Next() {
		err := rows.Scan(&success)
		if err != nil {
			logging.GlobalLogger.Err(err).Msg("")
			return false
		}
		break
	}
	if success == "1" {
		return true
	} else {
		return false
	}
}
