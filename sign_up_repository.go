package signup

import "context"

type SignUpRepository interface {
	CheckUserName(ctx context.Context, userName string) (bool, error)
	CheckUserNameAndContact(ctx context.Context, userName string, contact string) (bool, bool, error)
	Save(ctx context.Context, id string, user SignUpInfo) (bool, error)
	SavePasswordAndActivate(ctx context.Context, id, password string) (bool, error)
	SentVerifiedCode(ctx context.Context, id string) (bool, error)
	Activate(ctx context.Context, id string) (bool, error)
}
