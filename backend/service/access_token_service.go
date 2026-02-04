package service

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tfabritius/plainpage/service/ctxutil"
)

const accessTokenValidity = 15 * time.Minute // 15 minutes

func NewAccessTokenService(jwtSecret string) AccessTokenService {
	return AccessTokenService{
		jwtSecret: jwtSecret,
	}
}

type AccessTokenService struct {
	jwtSecret string
}

func (s *AccessTokenService) Create(userID string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": now.Unix(),
		"exp": now.Add(accessTokenValidity).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedTokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return signedTokenString, nil
}

func (s *AccessTokenService) validate(tokenString string) (string, error) {
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
		if sub, ok := claims["sub"].(string); ok {
			return sub, nil
		}
		return "", errors.New("invalid token: missing or invalid sub claim")
	}

	return "", errors.New("invalid token")
}

func (s *AccessTokenService) Token2ContextMiddleware(next http.Handler) http.Handler {
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

		id, err := s.validate(bearerToken[1])
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
