package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey struct{}

var claimsKey contextKey

// RequireAuth rejects requests without a valid session token and puts the
// verified claims in the request context for the handler.
func (s *Service) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := BearerToken(r)
		if token == "" {
			writeError(w, http.StatusUnauthorized, "missing session token")
			return
		}

		claims, err := s.Verify(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired session")
			return
		}

		next(w, r.WithContext(context.WithValue(r.Context(), claimsKey, claims)))
	}
}

// BearerToken extracts the token from an Authorization header.
func BearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	token, found := strings.CutPrefix(header, "Bearer ")
	if !found {
		return ""
	}
	return strings.TrimSpace(token)
}

// ClaimsFrom returns the claims RequireAuth stored on the request context.
func ClaimsFrom(ctx context.Context) (Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(Claims)
	return claims, ok
}
