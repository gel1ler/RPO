package httpapi

import (
	"net/http"
	"strings"

	"rpo/internal/auth"
)

type JWTParser interface {
	Parse(tokenString string) (auth.Claims, error)
}

func requireAuth(jwtParser JWTParser, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authz := r.Header.Get("Authorization")
		if authz == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized", "missing bearer token")
			return
		}
		if !strings.HasPrefix(authz, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "unauthorized", "missing bearer token")
			return
		}
		tokenString := strings.TrimSpace(strings.TrimPrefix(authz, "Bearer "))
		if tokenString == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized", "missing bearer token")
			return
		}

		claims, err := jwtParser.Parse(tokenString)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized", "invalid token")
			return
		}

		r = r.WithContext(withAuth(r.Context(), claims.UserID, claims.IsAdmin))
		next.ServeHTTP(w, r)
	})
}

func requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !currentIsAdmin(r.Context()) {
			writeError(w, http.StatusForbidden, "forbidden", "admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}
