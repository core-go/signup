package signup

type SignUpStatus int

const (
	StatusOK            = SignUpStatus(0)
	StatusUsernameError = SignUpStatus(1)
	StatusContactError  = SignUpStatus(2)
	StatusError         = SignUpStatus(4)
)
