package signup

import "context"

type SignUpService interface {
	SignUp(ctx context.Context, user SignUpInfo) (SignUpResult, error)
	VerifyUser(ctx context.Context, id, code string) (bool, error)
	VerifyUserAndSavePassword(ctx context.Context, id, code, password string) (int, error)
}
