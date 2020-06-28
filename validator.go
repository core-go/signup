package signup

import "context"

type Validator interface {
	Validate(ctx context.Context, user SignUpInfo) ([]ErrorMessage, error)
}
