package random

import (
	crand "crypto/rand"
	"math/big"
	mrand "math/rand"
)

const alphaNumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func AlphaNumeric(length int, needSecure bool) string {
	b := make([]byte, length)
	for i := range b {
		if needSecure {
			num, _ := crand.Int(crand.Reader, big.NewInt(int64(len(alphaNumeric))))
			b[i] = alphaNumeric[num.Int64()]
		} else {
			num := mrand.Intn(len(alphaNumeric))
			b[i] = alphaNumeric[num]
		}
	}

	return string(b)
}
