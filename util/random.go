package util

import (
	"math/rand"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const alphabet = "abcefghijklmnopqrstuvwxyz"

// RandomInt generate a random integer between min and max
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1) // min->max
}

// RandomString generate a random string of length n
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

// RandomOwner generate a random owner name
func RandomOwner() string {
	return RandomString(6)
}

// RandomMoney generate a random amount of money
func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

// RandomCurrency generate a random currency code
func RandomCurrency() string {
	currencies := []string{"USD", "EUR", "GBP", "VND"}
	n := len(currencies)
	return currencies[rand.Intn(n)]
}
