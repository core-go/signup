package signup

type SignUpActionConfig struct {
	Resource   string `mapstructure:"resource" json:"resource,omitempty" gorm:"column:resource" bson:"resource,omitempty" dynamodbav:"resource,omitempty" firestore:"resource,omitempty"`
	Signup     string `mapstructure:"sign_up" json:"signup,omitempty" gorm:"column:signup" bson:"signup,omitempty" dynamodbav:"signup,omitempty" firestore:"signup,omitempty"`
	VerifyUser string `mapstructure:"verify_user" json:"verifyUser,omitempty" gorm:"column:verifyuser" bson:"verifyUser,omitempty" dynamodbav:"verifyUser,omitempty" firestore:"verifyUser,omitempty"`
	Ip         string `mapstructure:"ip" json:"ip,omitempty" gorm:"column:ip" bson:"ip,omitempty" dynamodbav:"ip,omitempty" firestore:"ip,omitempty"`
}

type SignUpConfig struct {
	Expires      int                 `mapstructure:"expires" json:"expires,omitempty" gorm:"column:expires" bson:"expires,omitempty" dynamodbav:"expires,omitempty" firestore:"expires,omitempty"`
	Status       *SignUpStatusConfig `mapstructure:"status" json:"status,omitempty" gorm:"column:status" bson:"status,omitempty" dynamodbav:"status,omitempty" firestore:"status,omitempty"`
	UserStatus   UserStatusConf      `mapstructure:"user_status" json:"userStatus,omitempty" gorm:"column:userstatus" bson:"userStatus,omitempty" dynamodbav:"userStatus,omitempty" firestore:"userStatus,omitempty"`
	UserVerified *UserVerifiedConfig `mapstructure:"user_verified" json:"userVerified,omitempty" gorm:"column:userverified" bson:"userVerified,omitempty" dynamodbav:"userVerified,omitempty" firestore:"userVerified,omitempty"`
	Schema       *SignUpSchemaConfig `mapstructure:"schema" json:"schema,omitempty" gorm:"column:schema" bson:"schema,omitempty" dynamodbav:"schema,omitempty" firestore:"schema,omitempty"`
	Action       *SignUpActionConfig `mapstructure:"action" json:"action,omitempty" gorm:"column:action" bson:"action,omitempty" dynamodbav:"action,omitempty" firestore:"action,omitempty"`
}
