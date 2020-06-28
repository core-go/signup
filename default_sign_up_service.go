package signup

import (
	"context"
	"regexp"
	"time"
)

type DefaultSignUpService struct {
	UniqueContact          bool
	Repository             SignUpRepository
	IdGenerator            UserIdGenerator
	PasswordComparator     CodeComparator
	VerifiedCodeComparator CodeComparator
	VerifiedCodeRepository VerifiedCodeRepository
	VerifiedCodeSender     VerifiedCodeSender
	Expires                int
	Validator              Validator
	Regexps                []regexp.Regexp
	Generator              VerifiedCodeGenerator
}

func NewSignUpService(uniqueContact bool, repository SignUpRepository, idGenerator UserIdGenerator, passwordComparator CodeComparator, passcodeComparator CodeComparator, passcodeService VerifiedCodeRepository, passcodeSender VerifiedCodeSender, expires int, validator Validator, expressions []string, generator VerifiedCodeGenerator) *DefaultSignUpService {
	regExps := make([]regexp.Regexp, 0)
	if len(expressions) > 0 {
		for _, expression := range expressions {
			if len(expression) > 0 {
				regExp := regexp.MustCompile(expression)
				regExps = append(regExps, *regExp)
			}
		}
	}
	return &DefaultSignUpService{uniqueContact, repository, idGenerator, passwordComparator, passcodeComparator, passcodeService, passcodeSender, expires, validator, regExps, generator}
}

func (s *DefaultSignUpService) SignUp(ctx context.Context, user SignUpInfo) (SignUpResult, error) {
	result := SignUpResult{Status: StatusError}

	if s.Validator != nil {
		arrErr, er0 := s.Validator.Validate(ctx, user)
		if er0 != nil {
			return result, er0
		}
		if len(arrErr) > 0 {
			result.Errors = &arrErr
			return result, nil
		}
	}
	c := false
	var u bool
	var er1 error
	if s.UniqueContact {
		u, c, er1 = s.Repository.CheckUserNameAndContact(ctx, user.Username, user.Contact)
	} else {
		u, er1 = s.Repository.CheckUserName(ctx, user.Username)
	}
	if er1 != nil {
		return result, er1
	}
	if u {
		result.Status = StatusUsernameError
		return result, nil
	}
	if c {
		result.Status = StatusContactError
		return result, nil
	}

	hashPassword, er2 := s.PasswordComparator.Hash(user.Password)
	if er2 != nil {
		return result, nil
	}

	user.Password = hashPassword

	userId, er3 := s.IdGenerator.Generate(ctx)
	if er3 != nil {
		return result, er3
	}
	duplicate, er4 := s.Repository.Save(ctx, userId, user)
	if er4 != nil {
		return result, er4
	}

	if duplicate {
		i := 1
		for duplicate && i <= 5 {
			i++
			userId, er3 = s.IdGenerator.Generate(ctx)
			if er3 != nil {
				return result, er3
			}
			duplicate, er4 = s.Repository.Save(ctx, userId, user)
			if er4 != nil {
				return result, er4
			}
		}
		if duplicate {
			return result, nil
		}
	}

	_, er5 := s.createVerifiedCode(ctx, userId, user)
	if er5 != nil {
		return result, er1
	}

	result.Id = userId
	result.Status = StatusOK
	return result, nil
}

func (s *DefaultSignUpService) createVerifiedCode(ctx context.Context, userId string, user SignUpInfo) (bool, error) {
	verifiedCode := ""
	if s.Generator != nil {
		verifiedCode = s.Generator.Generate()
	} else {
		verifiedCode = generate(6)
	}

	hashedCode, er0 := s.VerifiedCodeComparator.Hash(verifiedCode)
	if er0 != nil {
		return false, er0
	}
	expiredAt := addSeconds(time.Now(), s.Expires)

	_, er1 := s.VerifiedCodeRepository.Save(ctx, userId, hashedCode, expiredAt)
	if er1 != nil {
		return false, er1
	}
	er2 := s.VerifiedCodeSender.Send(ctx, userId, verifiedCode, expiredAt, user.Contact)
	if er2 != nil {
		return false, er2
	}

	ok, er3 := s.Repository.SentVerifiedCode(ctx, userId)
	if ok && er3 == nil {
		return true, nil
	}
	return ok, er3
}

func (s *DefaultSignUpService) VerifyUser(ctx context.Context, userId string, code string) (bool, error) {
	ok1, er1 := s.verify(ctx, userId, code)
	if !ok1 {
		return ok1, er1
	}

	ok2, er2 := s.Repository.Activate(ctx, userId)
	if er2 == nil && ok2 {
		go func() {
			timeOut := 10 * time.Second
			ctxDelete, cancel := context.WithTimeout(ctx, timeOut)
			defer cancel()
			s.VerifiedCodeRepository.Delete(ctxDelete, userId)
		}()
	}
	return ok2, er2
}
func (s *DefaultSignUpService) VerifyUserAndSavePassword(ctx context.Context, userId, code, password string) (int, error) {
	if len(password) == 0 {
		return -1, nil
	} else if len(s.Regexps) > 0 {
		for _, exp := range s.Regexps {
			if !exp.MatchString(password) {
				return -1, nil
			}
		}
	}
	ok1, er1 := s.verify(ctx, userId, code)
	if !ok1 {
		return 0, er1
	}
	hashPassword, er1b := s.PasswordComparator.Hash(password)
	if er1b != nil {
		return 0, er1b
	}
	ok2, er2 := s.Repository.SavePasswordAndActivate(ctx, userId, hashPassword)
	if er2 == nil && ok2 {
		go func() {
			timeOut := 10 * time.Second
			ctxDelete, cancel := context.WithTimeout(ctx, timeOut)
			defer cancel()
			s.VerifiedCodeRepository.Delete(ctxDelete, userId)
		}()
	}
	if ok2 {
		return 1, er2
	}
	return 0, er2
}

func (s *DefaultSignUpService) verify(ctx context.Context, userId string, code string) (bool, error) {
	storedCode, expireAt, er0 := s.VerifiedCodeRepository.Load(ctx, userId)
	if er0 != nil {
		return false, er0
	}
	if time.Now().After(expireAt) {
		return false, nil
	}
	valid, er1 := s.VerifiedCodeComparator.Compare(code, storedCode)
	if !valid || er1 != nil {
		return false, er1
	}
	return true, nil
}

func addSeconds(date time.Time, seconds int) time.Time {
	return date.Add(time.Second * time.Duration(seconds))
}
