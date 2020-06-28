package signup

type VerifiedCodeGenerator interface {
	Generate() string
}
