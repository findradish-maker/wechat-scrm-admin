package jwt

import (
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

type Manager struct {
	secret      []byte
	expireHours int
}

type Claims struct {
	UserID   uint   `json:"userId"`
	Username string `json:"username"`
	Name     string `json:"name"`
	jwtv5.RegisteredClaims
}

func NewManager(secret string, expireHours int) *Manager {
	return &Manager{
		secret:      []byte(secret),
		expireHours: expireHours,
	}
}

func (m *Manager) Sign(userID uint, username, name string) (string, time.Time, error) {
	expiresAt := time.Now().Add(time.Duration(m.expireHours) * time.Hour)
	claims := Claims{
		UserID:   userID,
		Username: username,
		Name:     name,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(expiresAt),
			IssuedAt:  jwtv5.NewNumericDate(time.Now()),
			Subject:   username,
		},
	}
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	return signed, expiresAt, err
}

func (m *Manager) Parse(tokenString string) (*Claims, error) {
	token, err := jwtv5.ParseWithClaims(tokenString, &Claims{}, func(token *jwtv5.Token) (interface{}, error) {
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwtv5.ErrTokenInvalidClaims
	}
	return claims, nil
}
