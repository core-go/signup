package signup

type SignUpSchemaConfig struct {
	UserId   string `mapstructure:"user_id"`
	UserName string `mapstructure:"user_name"`
	Contact  string `mapstructure:"contact"`
	Password string `mapstructure:"password"`
	Status   string `mapstructure:"status"`

	SignedUpTime   string `mapstructure:"signed_up_time"`
	Language       string `mapstructure:"language"`
	MaxPasswordAge string `mapstructure:"max_password_age"`

	DateOfBirth string `mapstructure:"date_of_birth"`
	GivenName   string `mapstructure:"given_name"`
	MiddleName  string `mapstructure:"middle_name"`
	FamilyName  string `mapstructure:"family_name"`
	Gender      string `mapstructure:"gender"`

	CreatedTime string `mapstructure:"created_time"`
	CreatedBy   string `mapstructure:"created_by"`
	UpdatedTime string `mapstructure:"updated_time"`
	UpdatedBy   string `mapstructure:"updated_by"`
	Version     string `mapstructure:"version"`
}
