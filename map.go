package signup

import (
	"context"
	"time"
)

func BuildMap(ctx context.Context, user map[string]interface{}, userId string, info SignUpInfo, r SignUpSchemaConfig, genderMapper GenderMapper) map[string]interface{} {
	if len(info.Language) > 0 && len(r.Language) > 0 {
		user[r.Language] = info.Language
	}
	if info.DateOfBirth != nil && len(r.DateOfBirth) > 0 {
		user[r.DateOfBirth] = info.DateOfBirth
	}
	if len(info.GivenName) > 0 && len(r.GivenName) > 0 {
		user[r.GivenName] = info.GivenName
	}
	if len(info.MiddleName) > 0 && len(r.MiddleName) > 0 {
		user[r.MiddleName] = info.MiddleName
	}
	if len(info.FamilyName) > 0 && len(r.FamilyName) > 0 {
		user[r.FamilyName] = info.FamilyName
	}
	if len(info.Gender) > 0 && len(r.Gender) > 0 {
		if genderMapper != nil {
			g := genderMapper.Map(ctx, info.Gender)
			if g != nil {
				user[r.Gender] = g
			}
		} else {
			user[r.Gender] = info.Gender
		}
	}
	now := time.Now()
	if len(r.SignedUpTime) > 0 {
		user[r.SignedUpTime] = now
	}
	if len(r.CreatedTime) > 0 {
		user[r.CreatedTime] = now
	}
	if len(r.UpdatedTime) > 0 {
		user[r.UpdatedTime] = now
	}
	if len(r.CreatedBy) > 0 {
		user[r.CreatedBy] = userId
	}
	if len(r.UpdatedBy) > 0 {
		user[r.UpdatedBy] = userId
	}
	if len(r.Version) > 0 {
		user[r.Version] = 1
	}
	return user
}
