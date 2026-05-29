package httpapi

import (
	"net/http"
	"strings"

	"rpo/internal/auth"
)

type JWTParser interface {
	Parse(tokenString string) (auth.Claims, error)
	ParseTerminal(tokenString string) (auth.TerminalClaims, error)
}

func requireAuth(jwtParser JWTParser, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, ok := bearerTokenFromAuthHeader(r.Header.Get("Authorization"))
		if !ok {
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

// bearerTokenFromAuthHeader accepts "Bearer <jwt>" (рекомендуется) или голый JWT (как в Swagger UI).
func bearerTokenFromAuthHeader(authz string) (string, bool) {
	authz = strings.TrimSpace(authz)
	if authz == "" {
		return "", false
	}
	tokenString := authz
	for strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = strings.TrimSpace(strings.TrimPrefix(tokenString, "Bearer "))
	}
	if tokenString == "" {
		return "", false
	}
	return tokenString, true
}

func requireTerminalAuth(jwtParser JWTParser, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, ok := bearerTokenFromAuthHeader(r.Header.Get("Authorization"))
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized", "missing bearer token")
			return
		}

		claims, err := jwtParser.ParseTerminal(tokenString)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized", "invalid terminal token")
			return
		}

		r = r.WithContext(withTerminalAuth(r.Context(), claims.TerminalID, claims.SerialNumber))
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
