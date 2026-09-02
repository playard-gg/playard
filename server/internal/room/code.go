package room

import (
	"crypto/rand"
	"fmt"
	"strings"
)

// codeAlphabet excludes characters that are easy to misread or mistype when a
// code is spoken aloud or copied by hand: 0/O, 1/I/L, and the digit 8 vs B.
const codeAlphabet = "ACDEFGHJKMNPQRTUVWXY2346789"

// CodeLength gives ~27^6 ≈ 3.9e8 codes. Brute-forcing into someone else's
// room is prevented by rate limiting the join endpoint, not by length alone.
const CodeLength = 6

// NewCode generates a random, unbiased room code.
func NewCode() (string, error) {
	buf := make([]byte, CodeLength)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("room: generate code: %w", err)
	}

	// codeAlphabet's length divides evenly enough that modulo bias is
	// negligible here, but rejection sampling keeps it exactly uniform.
	var sb strings.Builder
	sb.Grow(CodeLength)
	limit := byte(256 - (256 % len(codeAlphabet)))
	for sb.Len() < CodeLength {
		for _, b := range buf {
			if b >= limit {
				continue
			}
			sb.WriteByte(codeAlphabet[int(b)%len(codeAlphabet)])
			if sb.Len() == CodeLength {
				break
			}
		}
		if sb.Len() < CodeLength {
			if _, err := rand.Read(buf); err != nil {
				return "", fmt.Errorf("room: generate code: %w", err)
			}
		}
	}
	return sb.String(), nil
}

// NormalizeCode makes user-typed codes forgiving: case-insensitive and
// tolerant of surrounding whitespace.
func NormalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}
