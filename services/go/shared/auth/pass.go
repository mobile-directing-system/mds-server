package auth

import "golang.org/x/crypto/bcrypt"

// BCryptHashCost is the cost to use when hashing via bcrypt.
const BCryptHashCost = bcrypt.DefaultCost
