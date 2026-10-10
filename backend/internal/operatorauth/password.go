package operatorauth

import "golang.org/x/crypto/bcrypt"

// bcryptCost is the library-recommended default. A human-chosen password comes
// from a comparatively small effective keyspace, which is exactly the case a
// slow, adaptive KDF exists to defend.
const bcryptCost = bcrypt.DefaultCost

// HashPassword one-way transforms a plaintext password for storage in
// operators.password_hash.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword reports whether password matches hash. bcrypt's comparison is
// already constant-time with respect to the password content.
func VerifyPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
