package signup

type SignUpStatusConfig struct {
	OK            *int `mapstructure:"ok" json:"ok,omitempty" gorm:"column:ok" bson:"ok,omitempty" dynamodbav:"ok,omitempty" firestore:"ok,omitempty"`
	UsernameError *int `mapstructure:"username" json:"username,omitempty" gorm:"column:username" bson:"username,omitempty" dynamodbav:"username,omitempty" firestore:"username,omitempty"`
	ContactError  *int `mapstructure:"contact" json:"contact,omitempty" gorm:"column:contact" bson:"contact,omitempty" dynamodbav:"contact,omitempty" firestore:"contact,omitempty"`
	Error         *int `mapstructure:"error" json:"error,omitempty" gorm:"column:error" bson:"error,omitempty" dynamodbav:"error,omitempty" firestore:"error,omitempty"`
}
type SignUpStatus struct {
	OK            int `mapstructure:"ok" json:"ok,omitempty" gorm:"column:ok" bson:"ok,omitempty" dynamodbav:"ok,omitempty" firestore:"ok,omitempty"`
	UsernameError int `mapstructure:"username" json:"username,omitempty" gorm:"column:username" bson:"username,omitempty" dynamodbav:"username,omitempty" firestore:"username,omitempty"`
	ContactError  int `mapstructure:"contact" json:"contact,omitempty" gorm:"column:contact" bson:"contact,omitempty" dynamodbav:"contact,omitempty" firestore:"contact,omitempty"`
	Error         int `mapstructure:"error" json:"error,omitempty" gorm:"column:error" bson:"error,omitempty" dynamodbav:"error,omitempty" firestore:"error,omitempty"`
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
