package mongo

import (
	"context"
	"fmt"
	"github.com/common-go/signup"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
	"time"
)

type SignUpRepository struct {
	UserCollection     *mongo.Collection
	PasswordCollection *mongo.Collection
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

func NewSignUpRepositoryByConfig(db *mongo.Database, userCollectionName, passwordCollectionName string, statusConfig signup.UserStatusConf, maxPasswordAge int, c *signup.SignUpSchemaConfig, options ...signup.GenderMapper) *SignUpRepository {
	var genderMapper signup.GenderMapper
	if len(options) > 0 {
		genderMapper = options[0]
	}
	userCollection := db.Collection(userCollectionName)
	passwordCollection := userCollection
	if passwordCollectionName != userCollectionName {
		passwordCollection = db.Collection(passwordCollectionName)
	}

	userName := c.UserName
	contact := c.Contact
	password := c.Password
	status := c.Status

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
		UserCollection:     userCollection,
		PasswordCollection: passwordCollection,
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

func NewSignUpRepository(db *mongo.Database, userCollectionName, passwordCollectionName string, statusConfig signup.UserStatusConf, maxPasswordAge int, maxPasswordAgeName string, userName, contactName, statusName string) *SignUpRepository {
	userCollection := db.Collection(userCollectionName)
	passwordCollection := userCollection
	if passwordCollectionName != userCollectionName {
		passwordCollection = db.Collection(passwordCollectionName)
	}
	if len(contactName) == 0 {
		contactName = "email"
	}
	return &SignUpRepository{
		UserCollection:     userCollection,
		PasswordCollection: passwordCollection,
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
	if r.Status.Registered == r.Status.Verifying {
		version = 2
	}
	return r.updateStatus(ctx, id, r.Status.Verifying, r.Status.Activated, version, "")
}

func (r *SignUpRepository) SentVerifiedCode(ctx context.Context, id string) (bool, error) {
	if r.Status.Registered == r.Status.Verifying {
		return true, nil
	}
	return r.updateStatus(ctx, id, r.Status.Registered, r.Status.Verifying, 2, "")
}

func (r *SignUpRepository) updateStatus(ctx context.Context, id string, from, to string, version int, password string) (bool, error) {
	query := bson.M{"$and": []bson.M{{"_id": id}, {r.StatusName: from}}}
	user := make(map[string]interface{})
	user["_id"] = id
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
	updateQuery := bson.M{
		"$set": user,
	}
	if len(password) > 0 && len(r.PasswordName) > 0 {
		user[r.PasswordName] = password
	}
	result, err := r.UserCollection.UpdateOne(ctx, query, updateQuery)
	return (result.ModifiedCount + result.UpsertedCount + result.MatchedCount) > 0, err
}

func (r *SignUpRepository) CheckUserName(ctx context.Context, userName string) (bool, error) {
	query := bson.M{r.UserName: userName}
	x := r.UserCollection.FindOne(ctx, query)
	err := x.Err()
	if err != nil {
		if fmt.Sprint(err) == "mongo: no documents in result" {
			return false, nil
		}
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
	query := bson.M{"$or": []bson.M{{r.UserName: userName}, {fieldName: fieldValue}}}

	findOptions := options.Find()
	findOptions.SetLimit(50)
	var fields = bson.M{}
	fields[r.UserName] = 1
	fields[fieldName] = 1

	cur, err := r.UserCollection.Find(ctx, query, findOptions)
	if err != nil {
		return false, false, err
	}
	nameErr := false
	contactErr := false
	for cur.Next(context.TODO()) {
		name := cur.Current.Lookup(r.UserName).StringValue()
		c := cur.Current.Lookup(fieldName).StringValue()
		if name == userName {
			nameErr = true
		}
		if c == fieldValue {
			contactErr = true
		}
	}
	//dont forget to close the cursor
	defer cur.Close(context.TODO())
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

	if r.UserCollection.Name() == r.PasswordCollection.Name() {
		user[r.PasswordName] = info.Password
	}
	duplicate, err := r.insertUser(ctx, userId, user)
	if err == nil && duplicate == false {
		if r.UserCollection.Name() != r.PasswordCollection.Name() {
			pass := make(map[string]interface{})
			pass["_id"] = userId
			pass[r.PasswordName] = info.Password
			_, er2 := r.PasswordCollection.InsertOne(ctx, pass)
			return false, er2
		}
	}
	return duplicate, err
}

func (r *SignUpRepository) SavePasswordAndActivate(ctx context.Context, userId, password string) (bool, error) {
	if r.UserCollection.Name() != r.PasswordCollection.Name() {
		user := make(map[string]interface{})
		user["_id"] = userId
		user[r.PasswordName] = password
		_, er2 := r.PasswordCollection.InsertOne(ctx, user)
		if er2 != nil {
			return false, er2
		}
		return r.Activate(ctx, userId)
	}

	version := 3
	if r.Status.Registered == r.Status.Verifying {
		version = 2
	}
	return r.updateStatus(ctx, userId, r.Status.Verifying, r.Status.Activated, version, password)
}

func (r *SignUpRepository) insertUser(ctx context.Context, userId string, user map[string]interface{}) (bool, error) {
	_, err := r.UserCollection.InsertOne(ctx, user)
	if err == nil {
		return false, nil
	}
	if strings.Index(err.Error(), "duplicate key error collection:") >= 0 {
		return true, nil
	} else {
		return false, err
	}
}
