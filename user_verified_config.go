package signup

type UserVerifiedConfig struct {
	Secure  bool   `mapstructure:"secure" json:"secure,omitempty" gorm:"column:secure" bson:"secure,omitempty" dynamodbav:"secure,omitempty" firestore:"secure,omitempty"`
	Domain  string `mapstructure:"domain" json:"domain,omitempty" gorm:"column:domain" bson:"domain,omitempty" dynamodbav:"domain,omitempty" firestore:"domain,omitempty"`
	Port    int    `mapstructure:"port" json:"port,omitempty" gorm:"column:port" bson:"port,omitempty" dynamodbav:"port,omitempty" firestore:"port,omitempty"`
	Action  string `mapstructure:"action" json:"action,omitempty" gorm:"column:action" bson:"action,omitempty" dynamodbav:"action,omitempty" firestore:"action,omitempty"`
}
