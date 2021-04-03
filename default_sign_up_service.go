package signup

import (
	"context"
	"regexp"
	"time"
)

type DefaultSignUpService struct {
	Status         SignUpStatus
	UniqueContact  bool
	Repository     SignUpRepository
	GenerateId     func(ctx context.Context) (string, error)
	Hash           func(plaintext string) (string, error)
	CodeComparator CodeComparator
	CodeRepository VerifiedCodeRepository
	SendCode       func(ctx context.Context, to string, code string, expireAt time.Time, params interface{}) error
	Expires        int
	Validate       func(ctx context.Context, user SignUpInfo) ([]ErrorMessage, error)
	Regexps        []regexp.Regexp
	GenerateCode   func() string
}

func NewSignUpService(status SignUpStatus, uniqueContact bool, repository SignUpRepository, generateId func(ctx context.Context) (string, error), hash func(string) (string, error), passcodeComparator CodeComparator, passcodeRepository VerifiedCodeRepository, sendCode func(context.Context, string, string, time.Time, interface{}) error, expires int, validate func(context.Context, SignUpInfo) ([]ErrorMessage, error), expressions []string, options ...func() string) *DefaultSignUpService {
	regExps := make([]regexp.Regexp, 0)
	if len(expressions) > 0 {
		for _, expression := range expressions {
			if len(expression) > 0 {
				regExp := regexp.MustCompile(expression)
				regExps = append(regExps, *regExp)
			}
		}
	}
	var generate func() string
	if len(options) >= 1 {
		generate = options[0]
	}
	return &DefaultSignUpService{Status: status, UniqueContact: uniqueContact, Repository: repository, GenerateId: generateId, Hash: hash, CodeComparator: passcodeComparator, CodeRepository: passcodeRepository, SendCode: sendCode, Expires: expires, Validate: validate, Regexps: regExps, GenerateCode: generate}
}

func (s *DefaultSignUpService) SignUp(ctx context.Context, user SignUpInfo) (SignUpResult, error) {
	result := SignUpResult{Status: s.Status.Error}

	if s.Validate != nil {
		arrErr, er0 := s.Validate(ctx, user)
		if er0 != nil {
			return result, er0
		}
		if len(arrErr) > 0 {
			result.Errors = arrErr
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
		result.Status = s.Status.UsernameError
		return result, nil
	}
	if c {
		result.Status = s.Status.ContactError
		return result, nil
	}

	hashPassword, er2 := s.Hash(user.Password)
	if er2 != nil {
		return result, er2
	}

	user.Password = hashPassword

	userId, er3 := s.GenerateId(ctx)
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
			userId, er3 = s.GenerateId(ctx)
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
		return result, er5
	}

	result.Id = userId
	result.Status = s.Status.OK
	return result, nil
}

func (s *DefaultSignUpService) createVerifiedCode(ctx context.Context, userId string, user SignUpInfo) (bool, error) {
	verifiedCode := ""
	if s.GenerateCode != nil {
		verifiedCode = s.GenerateCode()
	} else {
		verifiedCode = generate(6)
	}

	hashedCode, er0 := s.CodeComparator.Hash(verifiedCode)
	if er0 != nil {
		return false, er0
	}
	expiredAt := addSeconds(time.Now(), s.Expires)

	_, er1 := s.CodeRepository.Save(ctx, userId, hashedCode, expiredAt)
	if er1 != nil {
		return false, er1
	}
	er2 := s.SendCode(ctx, userId, verifiedCode, expiredAt, user.Contact)
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
			s.CodeRepository.Delete(ctxDelete, userId)
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
	ok1, er2 := s.verify(ctx, userId, code)
	if !ok1 {
		return 0, er2
	}
	hashPassword, er1b := s.Hash(password)
	if er1b != nil {
		return 0, er1b
	}
	ok2, er3 := s.Repository.SavePasswordAndActivate(ctx, userId, hashPassword)
	if er3 == nil && ok2 {
		go func() {
			timeOut := 10 * time.Second
			ctxDelete, cancel := context.WithTimeout(ctx, timeOut)
			defer cancel()
			s.CodeRepository.Delete(ctxDelete, userId)
		}()
	}
	if ok2 {
		return 1, er3
	}
	return 0, er3
}

func (s *DefaultSignUpService) verify(ctx context.Context, userId string, code string) (bool, error) {
	storedCode, expireAt, er0 := s.CodeRepository.Load(ctx, userId)
	if er0 != nil {
		return false, er0
	}
	if time.Now().After(expireAt) {
		return false, nil
	}
	valid, er1 := s.CodeComparator.Compare(code, storedCode)
	if !valid || er1 != nil {
		return false, er1
	}
	return true, nil
}

func addSeconds(date time.Time, seconds int) time.Time {
	return date.Add(time.Second * time.Duration(seconds))
}
