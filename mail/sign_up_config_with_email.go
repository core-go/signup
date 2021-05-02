package mail

import (
	"github.com/core-go/mail"
	"github.com/core-go/signup"
)

type SignUpConfigWithEmailTemplate struct {
	Expires      int                        `mapstructure:"expires"`
	Status       *signup.SignUpStatusConfig `mapstructure:"status"`
	UserStatus   signup.UserStatusConf      `mapstructure:"user_status"`
	UserVerified *signup.UserVerifiedConfig `mapstructure:"user_verified"`
	Schema       *signup.SignUpSchemaConfig `mapstructure:"schema"`
	Template     *mail.TemplateConfig       `mapstructure:"template"`
	Action       *signup.SignUpActionConfig `mapstructure:"action"`
}
