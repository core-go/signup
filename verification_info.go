package signup

type VerificationInfo struct {
	Id       string `mapstructure:"id" json:"id,omitempty" gorm:"column:id" bson:"id,omitempty" dynamodbav:"id,omitempty" firestore:"id,omitempty"`
	Password string `mapstructure:"password" json:"password,omitempty" gorm:"column:password" bson:"password,omitempty" dynamodbav:"password,omitempty" firestore:"password,omitempty"`
	Passcode string `mapstructure:"passcode" json:"passcode,omitempty" gorm:"column:passcode" bson:"passcode,omitempty" dynamodbav:"passcode,omitempty" firestore:"passcode,omitempty"`
}
