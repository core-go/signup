package signup

import "context"

type UserIdGenerator interface {
	Generate(ctx context.Context) (string, error)
}
