package service

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tfabritius/plainpage/model"
	"github.com/tfabritius/plainpage/service/ctxutil"
)

const tokenValidityDuration = 7 * 24 * time.Hour

func NewTokenService(jwtSecret string) TokenService {
	return TokenService{
		jwtSecret: jwtSecret,
	}
}

type TokenService struct {
	jwtSecret string
}

func (s *TokenService) GenerateToken(user model.User) (string, error) {
	claims := jwt.MapClaims{}
	claims["id"] = user.ID

	claims["exp"] = time.Now().Add(tokenValidityDuration).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedTokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return signedTokenString, nil
}

func (s *TokenService) validateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse token: %v", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["id"].(string), nil
	} else {
		return "", errors.New("invalid token")
	}
}

func (s *TokenService) Token2ContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")

		if header == "" {
			next.ServeHTTP(w, r)
			return
		}

		bearerToken := strings.Split(header, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			http.Error(w, "authorization header has wrong format, expected: Bearer <token>", http.StatusBadRequest)
			return
		}

		id, err := s.validateToken(bearerToken[1])
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// Inject username into request context
		ctx := r.Context()
		ctx = ctxutil.WithUserID(ctx, id)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
