package signup

type SignUpStatusConfig struct {
	OK            *int `mapstructure:"ok"`
	UsernameError *int `mapstructure:"username"`
	ContactError  *int `mapstructure:"contact"`
	Error         *int `mapstructure:"error"`
}
type SignUpStatus struct {
	OK            int `mapstructure:"ok"`
	UsernameError int `mapstructure:"username"`
	ContactError  int `mapstructure:"contact"`
	Error         int `mapstructure:"error"`
}

func InitSignUpStatus(c *SignUpStatusConfig) SignUpStatus {
	var c1 SignUpStatusConfig
	if c != nil {
		c1 = *c
	}
	var s SignUpStatus
	if c1.OK != nil {
		s.OK = *c1.OK
	} else {
		s.OK = 0
	}
	if c1.UsernameError != nil {
		s.UsernameError = *c1.UsernameError
	} else {
		s.UsernameError = 1
	}
	if c1.ContactError != nil {
		s.ContactError = *c1.ContactError
	} else {
		s.ContactError = 2
	}
	if c1.Error != nil {
		s.Error = *c1.Error
	} else {
		s.Error = 4
	}
	return s
}
