package sql

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/common-go/signup"
)

const (
	driverPostgres   = "postgres"
	driverMysql      = "mysql"
	driverMssql      = "mssql"
	driverOracle     = "oracle"
	driverSqlite3    = "sqlite3"
	driverNotSupport = "no support"
)

type SignUpRepository struct {
	DB                 *sql.DB
	UserTable          string
	PasswordTable      string
	Status             signup.UserStatusConf
	MaxPasswordAge     int
	MaxPasswordAgeName string

	UserIdName       string
	UserName         string
	ContactName      string
	StatusName       string
	PasswordName     string
	SignedUpTimeName string

	UpdatedTimeName string
	UpdatedByName   string
	VersionName     string

	GenderMapper signup.GenderMapper
	Schema       *signup.SignUpSchemaConfig

	Driver     string
	BuildParam func(i int) string
}

func NewSignUpRepositoryByConfig(db *sql.DB, userTable, passwordTable string, statusConfig signup.UserStatusConf, maxPasswordAge int, c *signup.SignUpSchemaConfig, options ...signup.GenderMapper) *SignUpRepository {
	if len(c.UserName) == 0 {
		c.UserName = "username"
	}
	if len(c.Contact) == 0 {
		c.Contact = "email"
	}
	if len(c.Password) == 0 {
		c.Password = "password"
	}
	if len(c.Status) == 0 {
		c.Status = "status"
	}
	var genderMapper signup.GenderMapper
	if len(options) > 0 {
		genderMapper = options[0]
	}
	c.UserId = strings.ToLower(c.UserId)
	c.UserName = strings.ToLower(c.UserName)
	c.Contact = strings.ToLower(c.Contact)
	c.Password = strings.ToLower(c.Password)
	c.Status = strings.ToLower(c.Status)
	c.SignedUpTime = strings.ToLower(c.SignedUpTime)
	c.Language = strings.ToLower(c.Language)
	c.MaxPasswordAge = strings.ToLower(c.MaxPasswordAge)
	c.DateOfBirth = strings.ToLower(c.DateOfBirth)
	c.GivenName = strings.ToLower(c.GivenName)
	c.MiddleName = strings.ToLower(c.MiddleName)
	c.FamilyName = strings.ToLower(c.FamilyName)
	c.Gender = strings.ToLower(c.Gender)
	c.CreatedTime = strings.ToLower(c.CreatedTime)
	c.CreatedBy = strings.ToLower(c.CreatedBy)
	c.UpdatedTime = strings.ToLower(c.UpdatedTime)
	c.UpdatedBy = strings.ToLower(c.UpdatedBy)
	c.Version = strings.ToLower(c.Version)

	userName := c.UserName
	contact := c.Contact
	password := c.Password
	status := c.Status
	build := getBuild(db)
	driver := getDriver(db)
	r := &SignUpRepository{
		DB:                 db,
		UserTable:          userTable,
		PasswordTable:      passwordTable,
		Status:             statusConfig,
		MaxPasswordAge:     maxPasswordAge,
		GenderMapper:       genderMapper,
		Schema:             c,
		MaxPasswordAgeName: c.MaxPasswordAge,
		UserIdName:         c.UserId,
		UserName:           userName,
		ContactName:        contact,
		PasswordName:       password,
		StatusName:         status,
		SignedUpTimeName:   c.SignedUpTime,

		UpdatedTimeName: c.UpdatedBy,
		UpdatedByName:   c.UpdatedBy,
		VersionName:     c.Version,
		BuildParam:      build,
		Driver:          driver,
	}
	return r
}

func NewSignUpRepository(db *sql.DB, userTable, passwordTable string, statusConfig signup.UserStatusConf, maxPasswordAge int, maxPasswordAgeName string, userId string, options ...string) *SignUpRepository {
	var contactName string
	if len(options) > 0 && len(options[0]) > 0 {
		contactName = options[0]
	}
	if len(contactName) == 0 {
		contactName = "email"
	}
	build := getBuild(db)
	driver := getDriver(db)
	return &SignUpRepository{
		DB:                 db,
		UserTable:          userTable,
		PasswordTable:      passwordTable,
		Status:             statusConfig,
		MaxPasswordAge:     maxPasswordAge,
		MaxPasswordAgeName: maxPasswordAgeName,
		UserIdName:         userId,
		UserName:           "username",
		ContactName:        contactName,
		PasswordName:       "password",
		StatusName:         "status",
		BuildParam:         build,
		Driver:             driver,
	}
}

func (s *SignUpRepository) Activate(ctx context.Context, id string) (bool, error) {
	version := 3
	if s.Status.Registered == s.Status.Verifying {
		version = 2
	}
	return s.updateStatus(ctx, id, s.Status.Verifying, s.Status.Activated, version, "")
}

func (s *SignUpRepository) SentVerifiedCode(ctx context.Context, id string) (bool, error) {
	if s.Status.Registered == s.Status.Verifying {
		return true, nil
	}
	return s.updateStatus(ctx, id, s.Status.Registered, s.Status.Verifying, 2, "")
}
func (s *SignUpRepository) CheckUserName(ctx context.Context, userName string) (bool, error) {
	query := fmt.Sprintf("Select %s from %s where %s = %s", s.UserName, s.UserTable, s.UserName, s.BuildParam(0))
	rows, err := s.DB.QueryContext(ctx, query, userName)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	exist := rows.Next()
	var username string
	if err := rows.Scan(&username); err != nil {
		return false, err
	}
	if !exist {
		return false, nil
	}
	return true, nil
}

func (s *SignUpRepository) CheckUserNameAndContact(ctx context.Context, userName string, contact string) (bool, bool, error) {
	return s.existUserNameAndField(ctx, userName, s.ContactName, contact)
}

func (s *SignUpRepository) existUserNameAndField(ctx context.Context, userName string, fieldName string, fieldValue string) (bool, bool, error) {
	query := fmt.Sprintf("select %s,%s from %s where %s = %s or %s = %s", s.UserName, fieldName, s.UserTable, s.UserName, s.BuildParam(0), fieldName, s.BuildParam(1))
	rows, err := s.DB.QueryContext(ctx, query, userName, fieldValue)
	if err != nil {
		return false, false, err
	}
	defer rows.Close()
	arr := make(map[string]interface{})
	columns := make([]interface{}, 2)
	columnPointers := make([]interface{}, 2)
	cols, _ := rows.Columns()
	nameExist := false
	emailExist := false
	for rows.Next() {
		for i, _ := range columns {
			columnPointers[i] = &columns[i]
		}

		if err1 := rows.Scan(columnPointers...); err1 != nil {
			return false, false, err1
		}

		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			arr[colName] = *val
		}
		if string(arr[s.UserName].([]byte)) == userName {
			nameExist = true
		}
		if string(arr[fieldName].([]byte)) == fieldValue {
			emailExist = true
		}
	}
	return nameExist, emailExist, rows.Err()
}

func (s *SignUpRepository) Save(ctx context.Context, userId string, info signup.SignUpInfo) (bool, error) {
	user := make(map[string]interface{})
	user[s.UserIdName] = userId
	user[s.UserName] = info.Username
	user[s.ContactName] = info.Contact
	user[s.StatusName] = s.Status.Registered
	if s.MaxPasswordAge > 0 && len(s.MaxPasswordAgeName) > 0 {
		user[s.MaxPasswordAgeName] = s.MaxPasswordAge
	}
	if s.Schema != nil {
		user = signup.BuildMap(ctx, user, userId, info, *s.Schema, s.GenderMapper)
	}

	tx, er0 := s.DB.Begin()
	if er0 != nil {
		return false, er0
	}
	if s.UserTable != s.PasswordTable && len(info.Password) > 0 {
		pass := make(map[string]interface{})
		pass[s.UserIdName] = userId
		pass[s.PasswordName] = info.Password
		query, value := BuildInsert(s.UserTable, user, s.BuildParam)
		passQuery, passValue := BuildInsert(s.PasswordTable, pass, s.BuildParam)
		_, er1 := tx.Exec(query, value...)
		if er1 != nil {
			tx.Rollback()
			return handleDuplicate(s.Driver, er1)
		}
		_, er2 := tx.Exec(passQuery, passValue...)
		if er2 != nil {
			tx.Rollback()
			return handleDuplicate(s.Driver, er2)
		}
		if er3 := tx.Commit(); er3 != nil {
			tx.Rollback()
			return false, er3
		}
		return false, nil
	}
	if len(info.Password) > 0 {
		user[s.PasswordName] = info.Password
	}
	query, value := BuildInsert(s.UserTable, user, s.BuildParam)
	_, err4 := tx.Exec(query, value...)
	if err4 != nil {
		tx.Rollback()
		return handleDuplicate(s.Driver, err4)
	}
	if err5 := tx.Commit(); err5 != nil {
		tx.Rollback()
		return false, err5
	}
	return true, nil
}

func (s *SignUpRepository) SavePasswordAndActivate(ctx context.Context, userId, password string) (bool, error) {
	user := make(map[string]interface{})
	user[s.UserIdName] = userId
	user[s.PasswordName] = password
	query, value := BuildInsert(s.PasswordTable, user, s.BuildParam)
	tx, er1 := s.DB.Begin()
	if er1 != nil {
		return false, er1
	}
	_, er2 := tx.Exec(query, value...)
	if er2 != nil {
		tx.Rollback()
		return handleDuplicate(s.Driver, er2)
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return s.Activate(ctx, userId)
}

func (s *SignUpRepository) updateStatus(ctx context.Context, id string, from, to string, version int, password string) (bool, error) {
	user := make(map[string]interface{})
	user[s.StatusName] = to
	if len(s.UpdatedTimeName) > 0 {
		user[s.UpdatedTimeName] = time.Now()
	}
	if len(s.UpdatedByName) > 0 {
		user[s.UpdatedByName] = id
	}
	if len(s.VersionName) > 0 && version > 0 {
		user[s.VersionName] = version
	}
	if s.UserTable == s.PasswordTable && len(password) > 0 && len(s.PasswordName) > 0 {
		user[s.PasswordName] = password
	}
	tx, err1 := s.DB.Begin()
	if err1 != nil {
		return false, err1
	}
	colNumber := 0
	var values []interface{}
	table := s.UserTable
	querySet := make([]string, 0)
	for colName, v2 := range user {
		values = append(values, v2)
		querySet = append(querySet, fmt.Sprintf("%v="+s.BuildParam(colNumber), colName))
		colNumber++
	}
	queryWhere := fmt.Sprintf(" %s = %s and %s = %s ",
		s.UserIdName,
		s.BuildParam(colNumber),
		s.StatusName,
		s.BuildParam(colNumber+1),
	)
	values = append(values, id)
	values = append(values, from)
	query := fmt.Sprintf("update %v set %v where %v", table, strings.Join(querySet, ","), queryWhere)
	result, err1 := s.DB.ExecContext(ctx, query, values...)
	if err1 != nil {
		tx.Rollback()
		return false, err1
	}
	if err2 := tx.Commit(); err2 != nil {
		tx.Rollback()
		return false, err2
	}
	r, err3 := result.RowsAffected()
	if err3 != nil {
		tx.Rollback()
		return false, err3
	}
	return r > 0, nil
}

func handleDuplicate(driver string, err error) (bool, error) {
	switch driver {
	case driverPostgres:
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			return true, nil
		}
		return false, err
	case driverMysql:
		if strings.Contains(err.Error(), "Error 1062: Duplicate entry") {
			return true, nil
		}
		return false, err
	case driverMssql:
		if strings.Contains(err.Error(), "Violation of PRIMARY KEY constraint") {
			return true, nil
		}
		return false, err
	case driverOracle:
		if strings.Contains(err.Error(), "ORA-00001: unique constraint") {
			return true, nil
		}
		return false, err
	case driverSqlite3:
		if strings.Contains(err.Error(), "UNIQUE constraint failed:") {
			return true, nil
		}
		return false, err
	default:
		return false, err
	}
}
func BuildInsert(tableName string, user map[string]interface{}, buildParam func(i int) string) (string, []interface{}) {
	var cols []string
	var values []interface{}
	for col, v := range user {
		cols = append(cols, col)
		values = append(values, v)
	}
	column := fmt.Sprintf("(%v)", strings.Join(cols, ","))
	numCol := len(cols)
	var arrValue []string
	for i := 0; i < numCol; i++ {
		arrValue = append(arrValue, buildParam(i))
	}
	value := fmt.Sprintf("(%v)", strings.Join(arrValue, ","))
	return fmt.Sprintf("INSERT INTO %v %v VALUES %v", tableName, column, value), values
}
func buildParam(i int) string {
	return "?"
}
func buildOracleParam(i int) string {
	return ":val" + strconv.Itoa(i)
}
func buildMsSqlParam(i int) string {
	return "@p" + strconv.Itoa(i)
}
func buildDollarParam(i int) string {
	return "$" + strconv.Itoa(i)
}
func getBuild(db *sql.DB) func(i int) string {
	driver := reflect.TypeOf(db.Driver()).String()
	switch driver {
	case "*pq.Driver":
		return buildDollarParam
	case "*godror.drv":
		return buildOracleParam
	case "*mssql.Driver":
		return buildMsSqlParam
	default:
		return buildParam
	}
}
func getDriver(db *sql.DB) string {
	if db == nil {
		return driverNotSupport
	}
	driver := reflect.TypeOf(db.Driver()).String()
	switch driver {
	case "*pq.Driver":
		return driverPostgres
	case "*godror.drv":
		return driverOracle
	case "*mysql.MySQLDriver":
		return driverMysql
	case "*mssql.Driver":
		return driverMssql
	case "*sqlite3.SQLiteDriver":
		return driverSqlite3
	default:
		return driverNotSupport
	}
}
