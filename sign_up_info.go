package signup

import "time"

type SignUpInfo struct {
	Username    string     `mapstructure:"username" json:"username,omitempty" gorm:"column:username" bson:"username,omitempty" dynamodbav:"username,omitempty" firestore:"username,omitempty"`
	Password    string     `mapstructure:"password" json:"password,omitempty" gorm:"column:password" bson:"password,omitempty" dynamodbav:"password,omitempty" firestore:"password,omitempty"`
	Contact     string     `mapstructure:"contact" json:"contact,omitempty" gorm:"column:contact" bson:"contact,omitempty" dynamodbav:"contact,omitempty" firestore:"contact,omitempty"`
	Email       string     `mapstructure:"email" json:"email,omitempty" gorm:"column:email" bson:"email,omitempty" dynamodbav:"email,omitempty" firestore:"email,omitempty"`
	Phone       string     `mapstructure:"phone" json:"phone,omitempty" gorm:"column:phone" bson:"phone,omitempty" dynamodbav:"phone,omitempty" firestore:"phone,omitempty"`
	Language    string     `mapstructure:"language" json:"language,omitempty" gorm:"column:language" bson:"language,omitempty" dynamodbav:"language,omitempty" firestore:"language,omitempty"`
	DateOfBirth *time.Time `mapstructure:"date_of_birth" json:"dateOfBirth,omitempty" gorm:"column:dateofbirth" bson:"dateOfBirth,omitempty" dynamodbav:"dateOfBirth,omitempty" firestore:"dateOfBirth,omitempty"`
	Gender      string     `mapstructure:"gender" json:"gender,omitempty" gorm:"column:gender" bson:"gender,omitempty" dynamodbav:"gender,omitempty" firestore:"gender,omitempty"`
	GivenName   string     `mapstructure:"given_name" json:"givenName,omitempty" gorm:"column:givenname" bson:"givenName,omitempty" dynamodbav:"givenName,omitempty" firestore:"givenName,omitempty"`
	MiddleName  string     `mapstructure:"middle_name" json:"middleName,omitempty" gorm:"column:middlename" bson:"middleName,omitempty" dynamodbav:"middleName,omitempty" firestore:"middleName,omitempty"`
	FamilyName  string     `mapstructure:"family_name" json:"familyName,omitempty" gorm:"column:familyname" bson:"familyName,omitempty" dynamodbav:"familyName,omitempty" firestore:"familyName,omitempty"`
}
