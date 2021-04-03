package signup

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	DriverPostgres   = "postgres"
	DriverMysql      = "mysql"
	DriverMssql      = "mssql"
	DriverOracle     = "oracle"
	DriverNotSupport = "no support"
)

type SqlSignUpRepository struct {
	DB                 *sql.DB
	UserTable          string
	PasswordTable      string
	Status             UserStatusConf
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

	GenderMapper GenderMapper
	Schema       *SignUpSchemaConfig

	Driver string
}

func NewSqlSignUpRepositoryByConfig(db *sql.DB, userTable, passwordTable string, statusConfig UserStatusConf, maxPasswordAge int, c *SignUpSchemaConfig, options...GenderMapper) *SqlSignUpRepository {
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
	var genderMapper GenderMapper
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

	r := &SqlSignUpRepository{
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
		Driver:          GetDriver(db),
	}
	return r
}

func NewSqlSignUpRepository(userTable, passwordTable string, statusConfig UserStatusConf, maxPasswordAge int, maxPasswordAgeName string, userId string, options...string) *SqlSignUpRepository {
	var contactName string
	if len(options) > 0 && len(options[0]) > 0 {
		contactName = options[0]
	}
	if len(contactName) == 0 {
		contactName = "email"
	}
	return &SqlSignUpRepository{
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
	}
}

func (s *SqlSignUpRepository) Activate(ctx context.Context, id string) (bool, error) {
	version := 3
	if s.Status.Registered == s.Status.Verifying {
		version = 2
	}
	return s.updateStatus(ctx, id, s.Status.Verifying, s.Status.Activated, version, "")
}

func (s *SqlSignUpRepository) SentVerifiedCode(ctx context.Context, id string) (bool, error) {
	if s.Status.Registered == s.Status.Verifying {
		return true, nil
	}
	return s.updateStatus(ctx, id, s.Status.Registered, s.Status.Verifying, 2, "")
}
func (s *SqlSignUpRepository) CheckUserName(ctx context.Context, userName string) (bool, error) {
	query := fmt.Sprintf("Select %s from %s where %s = %s", s.UserName, s.UserTable, s.UserName, BuildParam(0, s.Driver))
	rows, err := s.DB.Query(query, userName)
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

func (s *SqlSignUpRepository) CheckUserNameAndContact(ctx context.Context, userName string, contact string) (bool, bool, error) {
	return s.existUserNameAndField(ctx, userName, s.ContactName, contact)
}

func (s *SqlSignUpRepository) existUserNameAndField(ctx context.Context, userName string, fieldName string, fieldValue string) (bool, bool, error) {
	query := fmt.Sprintf("select %s,%s from %s where %s = %s or %s = %s", s.UserName, fieldName, s.UserTable, s.UserName, BuildParam(0, s.Driver), fieldName, BuildParam(1, s.Driver))
	rows, err := s.DB.Query(query, userName, fieldValue)
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

func (s *SqlSignUpRepository) Save(ctx context.Context, userId string, info SignUpInfo) (bool, error) {
	user := make(map[string]interface{})
	user[s.UserIdName] = userId
	user[s.UserName] = info.Username
	user[s.ContactName] = info.Contact
	user[s.StatusName] = s.Status.Registered
	if s.MaxPasswordAge > 0 && len(s.MaxPasswordAgeName) > 0 {
		user[s.MaxPasswordAgeName] = s.MaxPasswordAge
	}
	if s.Schema != nil {
		user = BuildMap(ctx, user, userId, info, *s.Schema, s.GenderMapper)
	}

	tx, err0 := s.DB.Begin()
	if err0 != nil {
		return false, err0
	}
	if s.UserTable != s.PasswordTable && len(info.Password) > 0 {
		pass := make(map[string]interface{})
		pass[s.UserIdName] = userId
		pass[s.PasswordName] = info.Password
		query, value := s.insertUser(s.UserTable, user, s.Driver)
		passQuery, passValue := s.insertUser(s.PasswordTable, pass, s.Driver)
		_, err1 := tx.Exec(query, value...)
		if err1 != nil {
			tx.Rollback()
			return s.handleDuplicate(err1)
		}
		_, err2 := tx.Exec(passQuery, passValue...)
		if err2 != nil {
			tx.Rollback()
			return s.handleDuplicate(err2)
		}
		if err3 := tx.Commit(); err3 != nil {
			tx.Rollback()
			return false, err3
		}
		return false, nil
	}
	if len(info.Password) > 0 {
		user[s.PasswordName] = info.Password
	}
	query, value := s.insertUser(s.UserTable, user, s.Driver)
	_, err4 := tx.Exec(query, value...)
	if err4 != nil {
		tx.Rollback()
		return s.handleDuplicate(err4)
	}
	if err5 := tx.Commit(); err5 != nil {
		tx.Rollback()
		return false, err5
	}
	return true, nil
}

func (s *SqlSignUpRepository) SavePasswordAndActivate(ctx context.Context, userId, password string) (bool, error) {
	user := make(map[string]interface{})
	user[s.UserIdName] = userId
	user[s.PasswordName] = password
	query, value := s.insertUser(s.PasswordTable, user, s.Driver)
	tx, err1 := s.DB.Begin()
	if err1 != nil {
		return false, err1
	}
	_, err2 := tx.Exec(query, value...)
	if err2 != nil {
		tx.Rollback()
		return s.handleDuplicate(err2)
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return s.Activate(ctx, userId)
}

func (s *SqlSignUpRepository) insertUser(tableName string, user map[string]interface{}, driverName string) (string, []interface{}) {
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
		arrValue = append(arrValue, BuildParam(i, driverName))
	}
	value := fmt.Sprintf("(%v)", strings.Join(arrValue, ","))
	return fmt.Sprintf("INSERT INTO %v %v VALUES %v", tableName, column, value), values
}

func (s *SqlSignUpRepository) updateStatus(ctx context.Context, id string, from, to string, version int, password string) (bool, error) {
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
	values := []interface{}{}
	table := s.UserTable
	querySet := make([]string, 0)
	for colName, v2 := range user {
		values = append(values, v2)
		querySet = append(querySet, fmt.Sprintf("%v="+BuildParam(colNumber, s.Driver), colName))
		colNumber++
	}
	queryWhere := fmt.Sprintf(" %s = %s and %s = %s ",
		s.UserIdName,
		BuildParam(colNumber, s.Driver),
		s.StatusName,
		BuildParam(colNumber+1, s.Driver),
	)
	values = append(values, id)
	values = append(values, from)
	query := fmt.Sprintf("update %v set %v where %v", table, strings.Join(querySet, ","), queryWhere)
	result, err1 := s.DB.Exec(query, values...)
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

func BuildParam(index int, driver string) string {
	switch driver {
	case DriverPostgres:
		return "$" + strconv.Itoa(index)
	case DriverOracle:
		return ":val" + strconv.Itoa(index)
	default:
		return "?"
	}
}

func (s *SqlSignUpRepository) handleDuplicate(err error) (bool, error) {
	switch dialect := s.Driver; dialect {
	case DriverPostgres:
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			return true, nil
		}
		return false, err
	case DriverMysql:
		if strings.Contains(err.Error(), "Error 1062: Duplicate entry") {
			return true, nil
		}
		return false, err
	case DriverMssql:
		if strings.Contains(err.Error(), "Violation of PRIMARY KEY constraint") {
			return true, nil
		}
		return false, err
	case DriverOracle:
		if strings.Contains(err.Error(), "ORA-00001: unique constraint") {
			return true, nil
		}
		return false, err
	case "sqlite3":
		if strings.Contains(err.Error(), "UNIQUE constraint failed:") {
			return true, nil
		}
		return false, err
	default:
		return false, err
	}
}

func GetDriver(db *sql.DB) string {
	if db == nil {
		return DriverNotSupport
	}
	driver := reflect.TypeOf(db.Driver()).String()
	switch driver {
	case "*pq.Driver":
		return DriverPostgres
	case "*mysql.MySQLDriver":
		return DriverMysql
	case "*mssql.Driver":
		return DriverMssql
	case "*godror.drv":
		return DriverOracle
	default:
		return DriverNotSupport
	}
}
