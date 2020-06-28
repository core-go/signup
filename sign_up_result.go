package signup

type SignUpResult struct {
	Id      string          `json:"id,omitempty" gorm:"column:id" bson:"_id,omitempty" dynamodbav:"id,omitempty" firestore:"id,omitempty"`
	Status  SignUpStatus    `json:"status" gorm:"column:status" bson:"status" dynamodbav:"status" firestore:"status"`
	Errors  *[]ErrorMessage `json:"errors,omitempty" gorm:"column:errors" bson:"errors,omitempty" dynamodbav:"errors,omitempty" firestore:"errors,omitempty"`
	Message string          `json:"message,omitempty" gorm:"column:message" bson:"message,omitempty" dynamodbav:"message,omitempty" firestore:"message,omitempty"`
}
