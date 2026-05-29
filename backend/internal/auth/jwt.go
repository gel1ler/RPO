package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	UserID  int64 `json:"user_id"`
	IsAdmin bool  `json:"is_admin"`
	jwt.RegisteredClaims
}

// TerminalClaims — JWT для NFC-терминала (отдельно от пользовательского).
type TerminalClaims struct {
	TerminalID   int64  `json:"terminal_id"`
	SerialNumber string `json:"serial_number"`
	jwt.RegisteredClaims
}

type JWT struct {
	Secret []byte
	Issuer string
	TTL    time.Duration
}

func (j JWT) Sign(userID int64, isAdmin bool, now time.Time) (string, error) {
	if len(j.Secret) == 0 {
		return "", errors.New("jwt secret is empty")
	}
	if j.TTL <= 0 {
		return "", errors.New("jwt ttl must be > 0")
	}

	claims := Claims{
		UserID:  userID,
		IsAdmin: isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.TTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.Secret)
}

func (j JWT) Parse(tokenString string) (Claims, error) {
	if len(j.Secret) == 0 {
		return Claims{}, errors.New("jwt secret is empty")
	}

	parsed, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return j.Secret, nil
	}, jwt.WithIssuer(j.Issuer))
	if err != nil {
		return Claims{}, ErrInvalidToken
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return Claims{}, ErrInvalidToken
	}

	if claims.UserID <= 0 {
		return Claims{}, ErrInvalidToken
	}

	return *claims, nil
}

func (j JWT) SignTerminal(terminalID int64, serialNumber string, now time.Time) (string, error) {
	if len(j.Secret) == 0 {
		return "", errors.New("jwt secret is empty")
	}
	if j.TTL <= 0 {
		return "", errors.New("jwt ttl must be > 0")
	}
	if terminalID <= 0 || serialNumber == "" {
		return "", errors.New("terminal id and serial are required")
	}

	claims := TerminalClaims{
		TerminalID:   terminalID,
		SerialNumber: serialNumber,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.Issuer,
			Subject:   "terminal",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.TTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.Secret)
}

func (j JWT) ParseTerminal(tokenString string) (TerminalClaims, error) {
	if len(j.Secret) == 0 {
		return TerminalClaims{}, errors.New("jwt secret is empty")
	}

	parsed, err := jwt.ParseWithClaims(tokenString, &TerminalClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return j.Secret, nil
	}, jwt.WithIssuer(j.Issuer))
	if err != nil {
		return TerminalClaims{}, ErrInvalidToken
	}

	claims, ok := parsed.Claims.(*TerminalClaims)
	if !ok || !parsed.Valid {
		return TerminalClaims{}, ErrInvalidToken
	}
	if claims.TerminalID <= 0 || claims.SerialNumber == "" || claims.Subject != "terminal" {
		return TerminalClaims{}, ErrInvalidToken
	}

	return *claims, nil
}
