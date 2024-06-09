package core

type Hasher interface {
	Hash(value string) (string, error)
	Compare(value string, hash string) bool
}
