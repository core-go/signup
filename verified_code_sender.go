package signup

import (
	"context"
	"time"
)

type VerifiedCodeSender interface {
	Send(ctx context.Context, to string, code string, expireAt time.Time, params interface{}) error
}
