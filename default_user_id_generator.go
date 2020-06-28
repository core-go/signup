package signup

import "context"

type DefaultUserIdGenerator struct {
	shortId    bool
}

func NewUserIdGenerator(shortId bool) *DefaultUserIdGenerator {
	generator := DefaultUserIdGenerator{shortId}
	return &generator
}

func (s *DefaultUserIdGenerator) Generate(ctx context.Context) (string, error) {
	if s.shortId {
		return shortId()
	}
	return randomId(), nil
}
