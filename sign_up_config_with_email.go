package signup

import "github.com/common-go/mail"

type SignUpConfigWithEmailTemplate struct {
	Expires      int                  `mapstructure:"expires"`
	Status       *SignUpStatusConfig  `mapstructure:"status"`
	UserStatus   UserStatusConf       `mapstructure:"user_status"`
	UserVerified *UserVerifiedConfig  `mapstructure:"user_verified"`
	Schema       *SignUpSchemaConfig  `mapstructure:"schema"`
	Template     *mail.TemplateConfig `mapstructure:"template"`
	Action       *SignUpActionConfig  `mapstructure:"action"`
}
