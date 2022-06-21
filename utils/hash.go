package utils

import (
	"crypto/sha512"
	"encoding/hex"
)

// allows to sha 512 hash `str` n `rounds` times
func sha512Hash(str string, rounds uint16) string {
	s := str
	for i := 1; i < int(rounds); i++ {
		h := sha512.New()
		h.Write([]byte(s))
		hash := hex.EncodeToString(h.Sum(nil))

		// update string
		s = hash
	}

	return s
}

// hash api key with sha512
func HashApiKey(clearTextApiKey string) string {
	return sha512Hash(clearTextApiKey, 3)
}
