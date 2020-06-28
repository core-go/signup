package signup

import (
	"context"
	"reflect"
	"regexp"
)

type PhoneValidator struct {
	PasswordRequired bool
	Regexp           *regexp.Regexp
	requiredFields   []string
	modelIndexes     map[string]int
}

func NewPhoneValidator(passwordRequired bool, paswordExp string, fields ...string) *PhoneValidator {
	requiredFields := []string{"Username", "Contact"}
	requiredFields = append(requiredFields, fields...)
	model := SignUpInfo{}
	modelIndexes := getModelIndexes(model)
	var regExp *regexp.Regexp = nil
	if passwordRequired && len(paswordExp) > 0 {
		regExp = regexp.MustCompile(paswordExp)
	}

	return &PhoneValidator{PasswordRequired: passwordRequired, Regexp: regExp, requiredFields: requiredFields, modelIndexes: modelIndexes}
}

func (v *PhoneValidator) Validate(ctx context.Context, user SignUpInfo) (msgs []ErrorMessage, err error) {
	valueOfModel := reflect.ValueOf(user)

	if requireMsg, err := RequireFields(valueOfModel, v.requiredFields, v.modelIndexes); err != nil {
		return nil, err
	} else {
		msgs = append(msgs, requireMsg...)
	}

	if len(user.Username) > 0 {
		userNameMsg := CheckUserName("username", user.Username)
		if userNameMsg != nil {
			msgs = append(msgs, *userNameMsg)
		}
	}

	if len(user.Contact) > 0 {
		emailMsg := CheckPhone("contact", user.Contact)
		if emailMsg != nil {
			msgs = append(msgs, *emailMsg)
		}
	}

	if v.PasswordRequired {
		msg := CheckRequired("password", user.Password)
		if msg != nil {
			msgs = append(msgs, *msg)
		} else if v.Regexp != nil {
			msg2 := CheckExpression("password", *v.Regexp, user.Password)
			msgs = append(msgs, *msg2)
		}
	}
	return msgs, err
}
