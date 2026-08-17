package auth

import "golang.org/x/crypto/bcrypt"

// PasswordService hashes and verifies passwords using bcrypt.
type PasswordService struct{}

func NewPasswordService() *PasswordService { return &PasswordService{} }

func (p *PasswordService) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (p *PasswordService) Verify(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
