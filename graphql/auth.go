package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type authClaims struct {
	AccountID string `json:"account_id"`
	Role      string `json:"role"`
	jwt.RegisteredClaims
}

type authContextKey struct{}

func issueToken(secret []byte, account *Account) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, authClaims{
		AccountID: account.ID,
		Role:      string(account.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(8 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}).SignedString(secret)
}

func authMiddleware(secret []byte, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if strings.HasPrefix(header, "Bearer ") {
			token, err := jwt.ParseWithClaims(strings.TrimPrefix(header, "Bearer "), &authClaims{},
				func(token *jwt.Token) (any, error) {
					if token.Method != jwt.SigningMethodHS256 {
						return nil, fmt.Errorf("unexpected signing method")
					}
					return secret, nil
				})
			if err == nil {
				if claims, ok := token.Claims.(*authClaims); ok && token.Valid {
					r = r.WithContext(context.WithValue(r.Context(), authContextKey{}, claims))
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func claimsFromContext(ctx context.Context) (*authClaims, error) {
	claims, ok := ctx.Value(authContextKey{}).(*authClaims)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}
	return claims, nil
}
