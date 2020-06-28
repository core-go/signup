package signup

type CodeComparator interface {
	Compare(plaintext string, hashed string) (bool, error)
	Hash(plaintext string) (string, error)
}

