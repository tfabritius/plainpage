package service

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tfabritius/plainpage/service/ctxutil"
	"github.com/tfabritius/plainpage/storage"
)

const jwtSecret = "testSecret"

func TestGenerateToken(t *testing.T) {
	r := require.New(t)

	tokenService := NewTokenService(jwtSecret)
	user := storage.User{ID: "test-user"}

	tokenString, err := tokenService.GenerateToken(user)
	r.NoError(err)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	r.NoError(err)
	r.True(token.Valid)

	claims, ok := token.Claims.(jwt.MapClaims)
	r.True(ok)
	r.Equal(user.ID, claims["id"].(string))

	exp, ok := claims["exp"].(float64)
	r.True(ok)
	r.LessOrEqual(time.Now().Unix(), int64(exp))
	r.Greater(int64(exp), time.Now().Add(14*time.Minute).Unix())
}

func TestToken2ContextMiddleware(t *testing.T) {
	r := require.New(t)

	tokenService := NewTokenService(jwtSecret)
	user := storage.User{ID: "test-user"}

	tokenString, err := tokenService.GenerateToken(user)
	r.NoError(err)

	claims := jwt.MapClaims{"id": user.ID, "exp": time.Now().Add(-1 * time.Second).Unix()}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	expiredTokenString, err := token.SignedString([]byte(jwtSecret))
	r.NoError(err)

	noUserIdHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := ctxutil.UserID(r.Context())
		assert.Empty(t, userID)
		w.WriteHeader(http.StatusOK)
	})

	checkUserIdHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := ctxutil.UserID(r.Context())
		assert.Equal(t, user.ID, userID)
		w.WriteHeader(http.StatusOK)
	})

	failHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.True(t, false)
	})

	testCases := []struct {
		name           string
		authHeader     string
		handler        http.HandlerFunc
		expectedStatus int
	}{
		{
			name:           "Valid token",
			authHeader:     fmt.Sprintf("Bearer %s", tokenString),
			handler:        checkUserIdHandler,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No token",
			authHeader:     "",
			handler:        noUserIdHandler,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Expired token",
			authHeader:     fmt.Sprintf("Bearer %s", expiredTokenString),
			handler:        failHandler,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid token",
			authHeader:     "Bearer invalid-token",
			handler:        failHandler,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Wrong format",
			authHeader:     fmt.Sprintf("Token %s", tokenString),
			handler:        failHandler,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://example.com", nil)
			req.Header.Set("Authorization", tc.authHeader)
			rec := httptest.NewRecorder()

			tokenService.Token2ContextMiddleware(tc.handler).ServeHTTP(rec, req)
			r.Equal(tc.expectedStatus, rec.Code)
		})
	}
}
