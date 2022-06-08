package auth

import (
	"github.com/lefinal/meh"
	"golang.org/x/crypto/bcrypt"
)

// BCryptHashCost is the cost to use when hashing via bcrypt.
const BCryptHashCost = bcrypt.DefaultCost

// PasswordOK checks if the given passwords match. The correct one is expected
// to be in hashed format.
func PasswordOK(hashedCorrectPassword []byte, testPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(hashedCorrectPassword, []byte(testPassword))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return false, meh.NewInternalErrFromErr(err, "compare hash and password", nil)
	}
	return true, nil
}

// HashPassword hashes the given password.
func HashPassword(pass string) ([]byte, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(pass), BCryptHashCost)
	if err != nil {
		return nil, meh.NewInternalErrFromErr(err, "hash password", nil)
	}
	return hashed, nil
}
