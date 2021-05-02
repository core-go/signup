package signup

type UserStatusConf struct {
	Registered string `mapstructure:"registered" json:"registered,omitempty" gorm:"column:registered" bson:"registered,omitempty" dynamodbav:"registered,omitempty" firestore:"registered,omitempty"`
	Verifying  string `mapstructure:"verifying" json:"verifying,omitempty" gorm:"column:verifying" bson:"verifying,omitempty" dynamodbav:"verifying,omitempty" firestore:"verifying,omitempty"`
	Activated  string `mapstructure:"activated" json:"activated,omitempty" gorm:"column:activated" bson:"activated,omitempty" dynamodbav:"activated,omitempty" firestore:"activated,omitempty"`
}
