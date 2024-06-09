package crypto

import "golang.org/x/crypto/bcrypt"

type BcryptHasher struct {
	cost int
}

func NewBcryptHasher(cost int) *BcryptHasher {
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}

	return &BcryptHasher{cost: cost}
}

func (h *BcryptHasher) Compare(value string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(value))
	return err == nil
}

func (h *BcryptHasher) Hash(value string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(value), h.cost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}
