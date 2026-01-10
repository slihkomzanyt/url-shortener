package random

import (
	crand "crypto/rand"
	"encoding/binary"
	"math/rand"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// NewAlias генерирует случайный алиас длиной n.
// Если n <= 0, вернёт пустую строку.
func NewAlias(n int) string {
	if n <= 0 {
		return ""
	}

	b := make([]byte, n)

	// Пытаемся криптостойко
	if _, err := crand.Read(b); err == nil {
		for i := 0; i < n; i++ {
			b[i] = alphabet[int(b[i])%len(alphabet)]
		}
		return string(b)
	}

	// Фолбэк: math/rand с seed из crypto/rand
	var seed int64
	_ = binary.Read(crand.Reader, binary.LittleEndian, &seed)
	r := rand.New(rand.NewSource(seed))

	for i := 0; i < n; i++ {
		b[i] = alphabet[r.Intn(len(alphabet))]
	}
	return string(b)
}
