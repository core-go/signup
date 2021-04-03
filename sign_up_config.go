package signup

type SignUpActionConfig struct {
	Resource   string `mapstructure:"resource"`
	Signup     string `mapstructure:"sign_up"`
	VerifyUser string `mapstructure:"verify_user"`
	Ip         string `mapstructure:"ip"`
}

type SignUpConfig struct {
	Expires      int                 `mapstructure:"expires"`
	Status       *SignUpStatusConfig `mapstructure:"status"`
	UserStatus   UserStatusConf      `mapstructure:"user_status"`
	UserVerified *UserVerifiedConfig `mapstructure:"user_verified"`
	Schema       *SignUpSchemaConfig `mapstructure:"schema"`
	Action       *SignUpActionConfig `mapstructure:"action"`
}
