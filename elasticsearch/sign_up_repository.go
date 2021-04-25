package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/elastic/go-elasticsearch"
	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/elastic/go-elasticsearch/esutil"

	db "github.com/common-go/elasticsearch"
	"github.com/common-go/signup"
)

type SignUpRepository struct {
	Client             *elasticsearch.Client
	UserIndexName      string
	PasswordIndexName  string
	Status             signup.UserStatusConf
	MaxPasswordAge     int
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

func NewSignUpRepositoryByConfig(db *elasticsearch.Client, userIndexName, passwordIndexName string, statusConfig signup.UserStatusConf, maxPasswordAge int, c *signup.SignUpSchemaConfig, options ...signup.GenderMapper) *SignUpRepository {
	var genderMapper signup.GenderMapper
	if len(options) > 0 {
		genderMapper = options[0]
	}
	userName := c.UserName
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

	return NewSignUpRepository(db, userIndexName, passwordIndexName, statusConfig, maxPasswordAge, c, genderMapper, userName, contact, password, status)
}

func NewSignUpRepository(db *elasticsearch.Client, userIndexName, passwordIndexName string, statusConfig signup.UserStatusConf, maxPasswordAge int, c *signup.SignUpSchemaConfig, genderMapper signup.GenderMapper, userName, contact, password, status string) *SignUpRepository {
	if len(contact) == 0 {
		contact = "email"
	}
	return &SignUpRepository{
		Client:             db,
		UserIndexName:      userIndexName,
		PasswordIndexName:  passwordIndexName,
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
	user := make(map[string]interface{})
	user[r.StatusName] = to
	if len(r.UpdatedTimeName) > 0 {
		user[r.UpdatedTimeName] = time.Now()
	}
	if len(r.UpdatedByName) > 0 {
		user[r.UpdatedByName] = id
	}
	if len(r.VersionName) > 0 && version > 0 {
		user[r.VersionName] = version
	}
	req := esapi.UpdateRequest{
		Index:      r.UserIndexName,
		DocumentID: id,
		Body:       esutil.NewJSONReader(user),
		Refresh:    "true",
	}
	res, err := req.Do(ctx, r.Client)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	if res.IsError() {
		return false, fmt.Errorf("document ID not exists in the index")
	}

	var temp map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&temp)
	if err != nil {
		return false, err
	}
	successful := int64(temp["_shards"].(map[string]interface{})["successful"].(float64))
	return successful > 0, nil
}

func (r *SignUpRepository) CheckUserName(ctx context.Context, userName string) (bool, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"match": map[string]interface{}{
					r.UserName: userName,
				},
			},
		},
	}
	res := make(map[string]interface{})
	ok, err := db.FindOneAndDecode(ctx, r.Client, []string{r.UserIndexName}, query, &res)
	if !ok || err != nil {
		return false, err
	}
	return true, nil
}

func (r *SignUpRepository) CheckUserNameAndContact(ctx context.Context, userName string, contact string) (bool, bool, error) {
	return r.existUserNameAndField(ctx, userName, r.ContactName, contact)
}

func (r *SignUpRepository) existUserNameAndField(ctx context.Context, userName string, fieldName string, fieldValue string) (bool, bool, error) {
	userName = strings.ToLower(userName)
	fieldValue = strings.ToLower(fieldValue)
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{r.UserName: userName}},
					{"term": map[string]interface{}{fieldName: fieldValue}},
				},
				"minimum_should_match": 1,
			},
		},
	}
	res := make([]map[string]interface{}, 0)
	ok, err := db.FindAndDecode(ctx, r.Client, []string{r.UserIndexName}, query, &res)
	if !ok || err != nil {
		return false, false, err
	}
	nameErr := false
	contactErr := false
	for i := range res {
		name := res[i][r.UserName].(string)
		c := res[i][fieldName].(string)
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
	user[r.UserName] = info.Username
	user[r.ContactName] = info.Contact
	user[r.StatusName] = r.Status.Registered
	if r.MaxPasswordAge > 0 && len(r.MaxPasswordAgeName) > 0 {
		user[r.MaxPasswordAgeName] = r.MaxPasswordAge
	}
	if r.Schema != nil {
		user = signup.BuildMap(ctx, user, userId, info, *r.Schema, r.GenderMapper)
	}

	if r.UserIndexName == r.PasswordIndexName {
		user[r.PasswordName] = info.Password
		return r.insertUser(ctx, userId, user)
	}
	duplicate, err := r.insertUser(ctx, userId, user)
	if err == nil && duplicate == false {
		pass := make(map[string]interface{})
		pass[r.PasswordName] = &dynamodb.AttributeValue{S: aws.String(info.Password)}
		req := esapi.CreateRequest{
			Index:      r.PasswordIndexName,
			DocumentID: userId,
			Body:       esutil.NewJSONReader(pass),
			Refresh:    "true",
		}
		res, err := req.Do(ctx, r.Client)
		if err != nil {
			return false, err
		}
		defer res.Body.Close()
		if res.IsError() {
			return true, fmt.Errorf("document ID already exists in the index")
		}
		var temp map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&temp); err != nil {
			return false, err
		}
	}
	return duplicate, err
}

func (r *SignUpRepository) SavePasswordAndActivate(ctx context.Context, userId, password string) (bool, error) {
	return r.Activate(ctx, userId)
}

func (r *SignUpRepository) insertUser(ctx context.Context, userId string, user map[string]interface{}) (bool, error) {
	req := esapi.CreateRequest{
		Index:      r.UserIndexName,
		DocumentID: userId,
		Body:       esutil.NewJSONReader(user),
		Refresh:    "true",
	}
	res, err := req.Do(ctx, r.Client)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	if res.IsError() {
		return true, fmt.Errorf("document ID already exists in the index")
	}
	var temp map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&temp); err != nil {
		return false, err
	}
	//fmt.Printf("[%s] %s; version=%d", res.Status(), temp["result"], int(temp["_version"].(float64)))
	return false, nil
}
