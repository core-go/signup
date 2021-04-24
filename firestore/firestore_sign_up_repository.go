package firestore

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/common-go/signup"
	"strings"
)

type FirestoreSignUpRepository struct {
	Client             *firestore.Client
	UserCollection     *firestore.CollectionRef
	PasswordCollection *firestore.CollectionRef
	Status             signup.UserStatusConf
	MaxPasswordAge     int
	MaxPasswordAgeName string

	UserName         string
	ContactName      string
	StatusName       string
	PasswordName     string
	SignedUpTimeName string

	UpdatedByName string
	VersionName   string

	GenderMapper signup.GenderMapper
	Schema       *signup.SignUpSchemaConfig
}

func NewSignUpRepositoryByConfig(client *firestore.Client, userCollectionName, passwordCollectionName string, statusConfig signup.UserStatusConf, maxPasswordAge int, c *signup.SignUpSchemaConfig, options ...signup.GenderMapper) *FirestoreSignUpRepository {
	var genderMapper signup.GenderMapper
	if len(options) > 0 {
		genderMapper = options[0]
	}
	userCollection := client.Collection(userCollectionName)
	passwordCollection := userCollection
	if passwordCollectionName != userCollectionName {
		passwordCollection = client.Collection(passwordCollectionName)
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

	r := &FirestoreSignUpRepository{
		Client:             client,
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

		UpdatedByName: c.UpdatedBy,
		VersionName:   c.Version,
	}
	return r
}

func NewSignUpRepository(client *firestore.Client, userCollectionName, passwordCollectionName string, statusConfig signup.UserStatusConf, maxPasswordAge int, maxPasswordAgeName string, userName, contactName string) *FirestoreSignUpRepository {
	userCollection := client.Collection(userCollectionName)
	passwordCollection := userCollection
	if passwordCollectionName != userCollectionName {
		passwordCollection = client.Collection(passwordCollectionName)
	}
	if len(contactName) == 0 {
		contactName = "email"
	}
	return &FirestoreSignUpRepository{
		Client:             client,
		UserCollection:     userCollection,
		PasswordCollection: passwordCollection,
		Status:             statusConfig,
		MaxPasswordAge:     maxPasswordAge,
		MaxPasswordAgeName: maxPasswordAgeName,
		UserName:           userName,
		ContactName:        contactName,
		PasswordName:       "password",
		StatusName:         "status",
	}
}

func (r *FirestoreSignUpRepository) Activate(ctx context.Context, id string) (bool, error) {
	version := 3
	if r.Status.Registered == r.Status.Verifying {
		version = 2
	}
	return r.updateStatus(ctx, id, r.Status.Verifying, r.Status.Activated, version)
}

func (r *FirestoreSignUpRepository) SentVerifiedCode(ctx context.Context, id string) (bool, error) {
	if r.Status.Registered == r.Status.Verifying {
		return true, nil
	}
	return r.updateStatus(ctx, id, r.Status.Registered, r.Status.Verifying, 2)
}

func (r *FirestoreSignUpRepository) updateStatus(ctx context.Context, id string, from, to string, version int) (bool, error) {
	err := r.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		docSnap, err := tx.Get(r.UserCollection.Doc(id))
		if err != nil {
			return err
		}
		if docSnap.Data()[r.StatusName] != from {
			return fmt.Errorf("invalid status")
		}
		updateValue := []firestore.Update{
			{Path: r.StatusName, Value: to},
		}
		if len(r.UpdatedByName) > 0 {
			updateValue = append(updateValue, firestore.Update{Path: r.UpdatedByName, Value: id})
		}
		if len(r.VersionName) > 0 && version > 0 {
			updateValue = append(updateValue, firestore.Update{Path: r.VersionName, Value: version})
		}
		return tx.Update(r.UserCollection.Doc(id), updateValue)
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *FirestoreSignUpRepository) CheckUserName(ctx context.Context, userName string) (bool, error) {
	_, err := r.UserCollection.Where(r.UserName, "=", userName).Limit(1).Documents(ctx).GetAll()
	if err != nil {
		if strings.Index(err.Error(), "Document already exists") >= 0 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *FirestoreSignUpRepository) CheckUserNameAndContact(ctx context.Context, userName string, contact string) (bool, bool, error) {
	return r.existUserNameAndField(ctx, userName, r.ContactName, contact)
}

func (r *FirestoreSignUpRepository) existUserNameAndField(ctx context.Context, userName string, fieldName string, fieldValue string) (bool, bool, error) {
	userName = strings.ToLower(userName)
	fieldValue = strings.ToLower(fieldValue)

	docs, err := r.UserCollection.Where(r.UserName, "==", userName).Documents(ctx).GetAll()
	if err != nil {
		return false, false, err
	}
	if len(docs) == 0 {
		docs, err := r.UserCollection.Where(fieldName, "==", fieldValue).Documents(ctx).GetAll()
		if err != nil {
			return false, false, err
		}
		if len(docs) == 0 {
			return false, false, err
		}
	}

	nameErr := false
	contactErr := false
	for _, doc := range docs {
		if nameErr && contactErr {
			break
		}
		mapValue := doc.Data()
		if mapValue[r.UserName] == userName {
			nameErr = true
		}
		if mapValue[fieldName] == fieldValue {
			contactErr = true
		}
	}

	return nameErr, contactErr, nil
}

func (r *FirestoreSignUpRepository) Save(ctx context.Context, userId string, info signup.SignUpInfo) (bool, error) {
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

	if r.UserCollection.ID == r.PasswordCollection.ID {
		user[r.PasswordName] = info.Password
	}
	delete(user, r.Schema.UpdatedTime)

	duplicate, err := r.insertUser(ctx, user, userId)
	if err == nil && duplicate == false {
		if r.UserCollection.ID != r.PasswordCollection.ID {
			pass := make(map[string]interface{})
			pass[r.PasswordName] = info.Password
			err := r.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
				return tx.Create(r.PasswordCollection.Doc(userId), pass)
			})
			return false, err
		}
	}
	return duplicate, err
}

func (r *FirestoreSignUpRepository) SavePasswordAndActivate(ctx context.Context, userId, password string) (bool, error) {
	pass := make(map[string]interface{})
	pass[r.PasswordName] = password
	err := r.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		return tx.Create(r.PasswordCollection.Doc(userId), pass)
	})
	if err != nil {
		return false, err
	}
	return r.Activate(ctx, userId)
}

func (r *FirestoreSignUpRepository) insertUser(ctx context.Context, user map[string]interface{}, id string) (bool, error) {
	err := r.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		return tx.Create(r.UserCollection.Doc(id), user)
	})
	if err == nil {
		return false, nil
	}
	if strings.Index(err.Error(), "Document already exists") >= 0 {
		return true, nil
	} else {
		return false, err
	}
}
