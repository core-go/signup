package dynamodb

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/core-go/signup"
	"strconv"
	"strings"
	"time"
)

type SignUpRepository struct {
	DB                 *dynamodb.DynamoDB
	UserTableName      string
	PasswordTableName  string
	MaxPasswordAge     int32
	Status             signup.UserStatusConf
	MaxPasswordAgeName string

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
}

func NewSignUpRepositoryByConfig(dynamoDB *dynamodb.DynamoDB, userTableName string, passwordTableName string, statusConfig signup.UserStatusConf, maxPasswordAge int32, c *signup.SignUpSchemaConfig, options ...signup.GenderMapper) *SignUpRepository {
	var genderMapper signup.GenderMapper
	if len(options) > 0 {
		genderMapper = options[0]
	}
	userName := c.Username
	contact := c.Contact
	password := c.Password
	status := c.Status
	/*
		if len(userId) == 0 {
			userId = "userId"
		}
	*/
	if len(userName) == 0 {
		userName = "userName"
	}
	if len(contact) == 0 {
		contact = "email"
	}
	if len(password) == 0 {
		password = "password"
	}
	if len(status) == 0 {
		status = "status"
	}

	r := &SignUpRepository{
		DB:                 dynamoDB,
		UserTableName:      userTableName,
		PasswordTableName:  passwordTableName,
		Status:             statusConfig,
		MaxPasswordAge:     maxPasswordAge,
		GenderMapper:       genderMapper,
		Schema:             c,
		MaxPasswordAgeName: c.MaxPasswordAge,
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

func NewSignUpRepository(dynamoDB *dynamodb.DynamoDB, userTableName string, passwordTableName string, statusConfig signup.UserStatusConf, maxPasswordAge int32, maxPasswordAgeName string, userName, contactName, statusName string) *SignUpRepository {
	if len(contactName) == 0 {
		contactName = "email"
	}
	return &SignUpRepository{
		DB:                 dynamoDB,
		UserTableName:      userTableName,
		PasswordTableName:  passwordTableName,
		Status:             statusConfig,
		MaxPasswordAge:     maxPasswordAge,
		MaxPasswordAgeName: maxPasswordAgeName,
		UserName:           userName,
		ContactName:        contactName,
		PasswordName:       "password",
		StatusName:         statusName,
	}
}

func (r *SignUpRepository) Activate(ctx context.Context, id string) (bool, error) {
	version := 3
	if strings.Compare(r.Status.Registered, r.Status.Verifying) == 0 {
		version = 2
	}
	return r.updateStatus(ctx, id, r.Status.Verifying, r.Status.Activated, version)
}

func (r *SignUpRepository) SentVerifiedCode(ctx context.Context, id string) (bool, error) {
	if strings.Compare(r.Status.Registered, r.Status.Verifying) == 0 {
		return true, nil
	}
	return r.updateStatus(ctx, id, r.Status.Registered, r.Status.Verifying, 2)
}

func (r *SignUpRepository) updateStatus(ctx context.Context, id string, from, to string, version int) (bool, error) {
	user := make(map[string]*dynamodb.AttributeValue)
	user["_id"] = &dynamodb.AttributeValue{S: aws.String(id)}
	user[r.StatusName] = &dynamodb.AttributeValue{S: aws.String(to)}
	if len(r.UpdatedTimeName) > 0 {
		user[r.UpdatedTimeName] = &dynamodb.AttributeValue{S: aws.String(time.Now().Format(time.RFC3339))}
	}
	if len(r.UpdatedByName) > 0 {
		user[r.UpdatedByName] = &dynamodb.AttributeValue{S: aws.String(id)}
	}
	if len(r.VersionName) > 0 && version > 0 {
		user[r.VersionName] = &dynamodb.AttributeValue{N: aws.String(strconv.Itoa(version))}
	}
	expected := make(map[string]*dynamodb.ExpectedAttributeValue)
	expected[r.StatusName] = &dynamodb.ExpectedAttributeValue{Value: &dynamodb.AttributeValue{S: aws.String(from)}, Exists: aws.Bool(true)}
	params := &dynamodb.PutItemInput{
		TableName:              aws.String(r.UserTableName),
		Expected:               expected,
		Item:                   user,
		ReturnConsumedCapacity: aws.String(dynamodb.ReturnConsumedCapacityTotal),
	}
	output, err := r.DB.PutItemWithContext(ctx, params)
	if err != nil {
		if strings.Index(err.Error(), "ConditionalCheckFailedException:") >= 0 {
			return false, fmt.Errorf("object not found")
		}
		return false, err
	}
	return int64(aws.Float64Value(output.ConsumedCapacity.CapacityUnits)) > 0, nil
}

func (r *SignUpRepository) CheckUserName(ctx context.Context, userName string) (bool, error) {
	filter := expression.Equal(expression.Name(r.UserName), expression.Value(userName))
	expr, _ := expression.NewBuilder().WithFilter(filter).Build()
	query := &dynamodb.ScanInput{
		TableName:                 aws.String(r.UserTableName),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}
	output, err := r.DB.ScanWithContext(ctx, query)
	if err != nil {
		return false, err
	}
	if len(output.Items) == 0 {
		return false, nil
	}
	return true, nil
}

func (r *SignUpRepository) CheckUserNameAndContact(ctx context.Context, userName string, contact string) (bool, bool, error) {
	return r.existUserNameAndField(ctx, userName, r.ContactName, contact)
}

func (r *SignUpRepository) existUserNameAndField(ctx context.Context, userName string, fieldName string, fieldValue string) (bool, bool, error) {
	userName = strings.ToLower(userName)
	fieldValue = strings.ToLower(fieldValue)
	userNameFilter := expression.Equal(expression.Name(r.UserName), expression.Value(userName))
	fieldNameFilter := expression.Equal(expression.Name(fieldName), expression.Value(fieldValue))
	projection := expression.NamesList(expression.Name("_id"), expression.Name(r.UserName), expression.Name(fieldName))
	filter := expression.Or(userNameFilter, fieldNameFilter)
	expr, _ := expression.NewBuilder().WithProjection(projection).WithFilter(filter).Build()
	query := &dynamodb.ScanInput{
		TableName:                 aws.String(r.UserTableName),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Limit:                     aws.Int64(50),
	}

	output, err := r.DB.ScanWithContext(ctx, query)
	if err != nil {
		return false, false, err
	}
	nameErr := false
	contactErr := false
	for i := range output.Items {
		name := strings.Trim(output.Items[i][r.UserName].String(), "\"")
		c := strings.Trim(output.Items[i][fieldName].String(), "\"")
		if name == userName {
			nameErr = true
		}
		if c == fieldValue {
			contactErr = true
		}
	}
	return nameErr, contactErr, nil
}

func (r *SignUpRepository) Save(ctx context.Context, userId string, info signup.SignUpInfo) (bool, error) {
	user := make(map[string]interface{})
	user["_id"] = userId
	user[r.UserName] = info.Username
	user[r.ContactName] = info.Contact
	user[r.StatusName] = r.Status.Registered
	if r.MaxPasswordAge > 0 && len(r.MaxPasswordAgeName) > 0 {
		user[r.MaxPasswordAgeName] = r.MaxPasswordAge
	}
	if r.Schema != nil {
		user = signup.BuildMap(ctx, user, userId, info, *r.Schema, r.GenderMapper)
	}

	if r.UserTableName == r.PasswordTableName {
		user[r.PasswordName] = info.Password
		return r.insertUser(ctx, userId, user)
	}

	duplicate, err := r.insertUser(ctx, userId, user)
	if err == nil && duplicate == false {
		pass := make(map[string]*dynamodb.AttributeValue)
		pass["_id"] = &dynamodb.AttributeValue{S: aws.String(userId)}
		pass[r.PasswordName] = &dynamodb.AttributeValue{S: aws.String(info.Password)}
		expected := make(map[string]*dynamodb.ExpectedAttributeValue)
		expected["_id"] = &dynamodb.ExpectedAttributeValue{Value: &dynamodb.AttributeValue{S: aws.String(userId)}, Exists: aws.Bool(false)}
		params := &dynamodb.PutItemInput{
			TableName:              aws.String(r.PasswordTableName),
			Expected:               expected,
			Item:                   pass,
			ReturnConsumedCapacity: aws.String(dynamodb.ReturnConsumedCapacityTotal),
		}
		_, err := r.DB.PutItemWithContext(ctx, params)
		if err == nil {
			return false, nil
		}
		if strings.Index(err.Error(), "ConditionalCheckFailedException:") >= 0 {
			return true, nil
		}
		return false, err
	}
	return duplicate, err
}

func (r *SignUpRepository) SavePasswordAndActivate(ctx context.Context, userId, password string) (bool, error) {
	user := make(map[string]interface{})
	user["_id"] = userId
	user[r.PasswordName] = password
	_, err := r.insertUser(ctx, userId, user)
	if err != nil {
		return false, err
	}
	return r.Activate(ctx, userId)
}

func (r *SignUpRepository) insertUser(ctx context.Context, userId string, user map[string]interface{}) (bool, error) {
	modelMap, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		return false, err
	}
	expected := make(map[string]*dynamodb.ExpectedAttributeValue)
	expected["_id"] = &dynamodb.ExpectedAttributeValue{Value: &dynamodb.AttributeValue{S: aws.String(userId)}, Exists: aws.Bool(false)}
	params := &dynamodb.PutItemInput{
		TableName:              aws.String(r.PasswordTableName),
		Expected:               expected,
		Item:                   modelMap,
		ReturnConsumedCapacity: aws.String(dynamodb.ReturnConsumedCapacityTotal),
	}
	_, err = r.DB.PutItemWithContext(ctx, params)
	if err != nil {
		if strings.Index(err.Error(), "ConditionalCheckFailedException:") >= 0 {
			return true, nil
		}
		return false, err
	}
	return false, nil
}
