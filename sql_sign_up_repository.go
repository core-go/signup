package signup

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

type SqlSignUpRepository struct {
	DB                 *gorm.DB
	UserTable          string
	PasswordTable      string
	Status             UserStatusConfig
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
}

func NewSqlSignUpRepositoryByConfig(db *gorm.DB, userTable, passwordTable string, statusConfig UserStatusConfig, maxPasswordAge int, c *SignUpSchemaConfig, genderMapper GenderMapper) *SqlSignUpRepository {

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
	}
	return r
}

func NewSqlSignUpRepository(userTable, passwordTable string, statusConfig UserStatusConfig, maxPasswordAge int, maxPasswordAgeName string, userId, contactName string) *SqlSignUpRepository {
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
	subScope := s.DB.NewScope("")
	query := fmt.Sprintf("Select %s from %s where %s = ?", subScope.Quote(s.UserName), subScope.Quote(s.UserTable), subScope.Quote(s.UserName))
	rows, err := s.DB.Raw(query, userName).Rows()
	if err != nil {
		return false, err
	}
	defer rows.Close()
	exist := rows.Next()
	if rows.Err() != nil {
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

	subScope := s.DB.NewScope("")
	query := fmt.Sprintf("Select %s,%s from %s where %s = ? or %s = ?", subScope.Quote(s.UserName), subScope.Quote(fieldName), subScope.Quote(s.UserTable), subScope.Quote(s.UserName), subScope.Quote(fieldName))
	rows, err := s.DB.Raw(query, userName, fieldValue).Rows()
	if err != nil {
		return false, false, err
	}
	defer rows.Close()
	exist := rows.Next()

	if !exist {
		return false, false, nil
	}
	arr := make(map[string]interface{})
	columns := make([]interface{}, 2)
	columnPointers := make([]interface{}, 2)
	if err1 := rows.Scan(columnPointers...); err1 != nil {
		return false, false, err1
	}
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
		} else if string(arr[fieldName].([]byte)) == fieldValue {
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

	tx := s.DB.Begin()
	if tx.Error != nil {
		return false, tx.Error
	}
	if s.UserTable != s.PasswordTable && len(info.Password) > 0 {
		pass := make(map[string]interface{})
		pass[s.UserIdName] = userId
		pass[s.PasswordName] = info.Password
		query, value := s.insertUser(s.UserTable, user)
		passQuery, passValue := s.insertUser(s.PasswordTable, pass)
		result1 := tx.Exec(query, value...)
		if err := result1.Error; err != nil {
			tx.Rollback()
			return s.handleDuplicate(err)
		}
		result2 := tx.Exec(passQuery, passValue...)
		if err := result2.Error; err != nil {
			tx.Rollback()
			return s.handleDuplicate(err)
		}
		return false, tx.Commit().Error
	}
	if len(info.Password) > 0 {
		user[s.PasswordName] = info.Password
	}
	query, value := s.insertUser(s.UserTable, user)
	result := tx.Exec(query, value...)
	if err := result.Error; err != nil {
		tx.Rollback()
		return s.handleDuplicate(err)
	}
	return false, tx.Commit().Error
}

func (s *SqlSignUpRepository) SavePasswordAndActivate(ctx context.Context, userId, password string) (bool, error) {
	user := make(map[string]interface{})
	user[s.UserIdName] = userId
	user[s.PasswordName] = password
	query, value := s.insertUser(s.PasswordTable, user)
	tx := s.DB.Begin()
	if tx.Error != nil {
		return false, tx.Error
	}
	result := tx.Exec(query, value...)
	if err := result.Error; err != nil {
		tx.Rollback()
		return s.handleDuplicate(err)
	}
	if err := tx.Commit().Error; err != nil {
		return false, err
	}
	return s.Activate(ctx, userId)
}

func (s *SqlSignUpRepository) insertUser(tableName string, user map[string]interface{}) (string, []interface{}) {
	var cols []string
	var values []interface{}
	subScope := s.DB.NewScope("")
	for col, v := range user {
		cols = append(cols, subScope.Quote(col))
		values = append(values, v)
	}
	column := fmt.Sprintf("(%v)", strings.Join(cols, ","))
	numCol := len(cols)
	var arrValue []string
	for i := 0; i < numCol; i++ {
		arrValue = append(arrValue, "?")
	}
	value := fmt.Sprintf("(%v)", strings.Join(arrValue, ","))
	return fmt.Sprintf("INSERT INTO %v %v VALUES %v", subScope.Quote(tableName), column, value), values
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
	scope := s.DB.NewScope("")
	tx := s.DB.Begin()
	if tx.Error != nil {
		return false, tx.Error
	}
	result := tx.Table(s.UserTable).Where(scope.Quote(s.UserIdName)+" = ? AND "+scope.Quote(s.StatusName)+" = ?", id, from).Updates(user)
	if err := result.Error; err != nil {
		tx.Rollback()
		return false, err
	}
	return result.RowsAffected > 0, tx.Commit().Error
}

func (s *SqlSignUpRepository) handleDuplicate(err error) (bool, error) {
	switch dialect := s.DB.Dialect().GetName(); dialect {
	case "postgres":
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
			return true, nil
		}
		return false, err
	case "mysql":
		if strings.Contains(err.Error(), "Error 1062: Duplicate entry") {
			return true, nil
		}
		return false, err
	case "mssql":
		if strings.Contains(err.Error(), "Violation of PRIMARY KEY constraint") {
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
