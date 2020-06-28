package signup

type SignUpConfig struct {
	Expires      int                 `mapstructure:"expires"`
	Status       UserStatusConfig    `mapstructure:"status"`
	UserVerified *UserVerifiedConfig `mapstructure:"user_verified"`
	Schema       *SignUpSchemaConfig `mapstructure:"schema"`
}
