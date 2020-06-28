package signup

import "context"

type GenderMapper interface {
	Map(ctx context.Context, gender string) interface{}
}
