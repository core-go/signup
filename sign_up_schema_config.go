package signup

type SignUpSchemaConfig struct {
	UserId   string `mapstructure:"user_id" json:"userId,omitempty" gorm:"column:userid" bson:"userId,omitempty" dynamodbav:"userId,omitempty" firestore:"userId,omitempty"`
	UserName string `mapstructure:"user_name" json:"username,omitempty" gorm:"column:username" bson:"username,omitempty" dynamodbav:"username,omitempty" firestore:"username,omitempty"`
	Contact  string `mapstructure:"contact" json:"contact,omitempty" gorm:"column:contact" bson:"contact,omitempty" dynamodbav:"contact,omitempty" firestore:"contact,omitempty"`
	Password string `mapstructure:"password" json:"password,omitempty" gorm:"column:password" bson:"password,omitempty" dynamodbav:"password,omitempty" firestore:"password,omitempty"`
	Status   string `mapstructure:"status" json:"status,omitempty" gorm:"column:status" bson:"status,omitempty" dynamodbav:"status,omitempty" firestore:"status,omitempty"`

	SignedUpTime   string `mapstructure:"signed_up_time" json:"signedUpTime,omitempty" gorm:"column:signeduptime" bson:"signedUpTime,omitempty" dynamodbav:"signedUpTime,omitempty" firestore:"signedUpTime,omitempty"`
	Language       string `mapstructure:"language" json:"language,omitempty" gorm:"column:language" bson:"language,omitempty" dynamodbav:"language,omitempty" firestore:"language,omitempty"`
	MaxPasswordAge string `mapstructure:"max_password_age" json:"maxPasswordAge,omitempty" gorm:"column:maxPasswordAge" bson:"maxPasswordAge,omitempty" dynamodbav:"maxPasswordAge,omitempty" firestore:"maxPasswordAge,omitempty"`

	DateOfBirth string `mapstructure:"date_of_birth" json:"dateOfBirth,omitempty" gorm:"column:dateOfBirth" bson:"dateOfBirth,omitempty" dynamodbav:"dateOfBirth,omitempty" firestore:"dateOfBirth,omitempty"`
	GivenName   string `mapstructure:"given_name" json:"givenName,omitempty" gorm:"column:givenname" bson:"givenName,omitempty" dynamodbav:"givenName,omitempty" firestore:"givenName,omitempty"`
	MiddleName  string `mapstructure:"middle_name" json:"middleName,omitempty" gorm:"column:middlename" bson:"middleName,omitempty" dynamodbav:"middleName,omitempty" firestore:"middleName,omitempty"`
	FamilyName  string `mapstructure:"family_name" json:"familyName,omitempty" gorm:"column:familyname" bson:"familyName,omitempty" dynamodbav:"familyName,omitempty" firestore:"familyName,omitempty"`
	Gender      string `mapstructure:"gender" json:"gender,omitempty" gorm:"column:gender" bson:"gender,omitempty" dynamodbav:"gender,omitempty" firestore:"gender,omitempty"`

	CreatedTime string `mapstructure:"created_time" json:"createdTime,omitempty" gorm:"column:createdtime" bson:"createdTime,omitempty" dynamodbav:"createdTime,omitempty" firestore:"createdTime,omitempty"`
	CreatedBy   string `mapstructure:"created_by" json:"createdBy,omitempty" gorm:"column:createdby" bson:"createdBy,omitempty" dynamodbav:"createdBy,omitempty" firestore:"createdBy,omitempty"`
	UpdatedTime string `mapstructure:"updated_time" json:"updatedTime,omitempty" gorm:"column:updatedtime" bson:"updatedTime,omitempty" dynamodbav:"updatedTime,omitempty" firestore:"updatedTime,omitempty"`
	UpdatedBy   string `mapstructure:"updated_by" json:"updatedBy,omitempty" gorm:"column:updatedby" bson:"updatedBy,omitempty" dynamodbav:"updatedBy,omitempty" firestore:"updatedBy,omitempty"`
	Version     string `mapstructure:"version" json:"version,omitempty" gorm:"column:version" bson:"version,omitempty" dynamodbav:"version,omitempty" firestore:"version,omitempty"`
}
