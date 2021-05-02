package mail

import (
	"github.com/core-go/mail"
	"github.com/core-go/signup"
)

type SignUpConfigWithEmailTemplate struct {
	Expires      int                        `mapstructure:"expires" json:"expires,omitempty" gorm:"column:expires" bson:"expires,omitempty" dynamodbav:"expires,omitempty" firestore:"expires,omitempty"`
	Status       *signup.SignUpStatusConfig `mapstructure:"status" json:"status,omitempty" gorm:"column:status" bson:"status,omitempty" dynamodbav:"status,omitempty" firestore:"status,omitempty"`
	UserStatus   signup.UserStatusConf      `mapstructure:"user_status" json:"userStatus,omitempty" gorm:"column:userstatus" bson:"userStatus,omitempty" dynamodbav:"userStatus,omitempty" firestore:"userStatus,omitempty"`
	UserVerified *signup.UserVerifiedConfig `mapstructure:"user_verified" json:"userVerified,omitempty" gorm:"column:userverified" bson:"userVerified,omitempty" dynamodbav:"userVerified,omitempty" firestore:"userVerified,omitempty"`
	Schema       *signup.SignUpSchemaConfig `mapstructure:"schema" json:"schema,omitempty" gorm:"column:schema" bson:"schema,omitempty" dynamodbav:"schema,omitempty" firestore:"schema,omitempty"`
	Template     *mail.TemplateConfig       `mapstructure:"template" json:"template,omitempty" gorm:"column:template" bson:"template,omitempty" dynamodbav:"template,omitempty" firestore:"template,omitempty"`
	Action       *signup.SignUpActionConfig `mapstructure:"action" json:"action,omitempty" gorm:"column:action" bson:"action,omitempty" dynamodbav:"action,omitempty" firestore:"action,omitempty"`
}
