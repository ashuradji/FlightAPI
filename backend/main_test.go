package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	r.POST("/login", LoginHandler)

	protected := r.Group("/secret")
	protected.Use(JWTAuthMiddleware())
	{
		protected.GET("/", JWTAuthMiddleware())
	}
	return r
}

func TestLogin(t *testing.T) {
	router := setupRouter()

	tests := []struct {
		name           string
		body           map[string]string
		expectedStatus int
		expectToken    bool
	}{
		{
			name: "valid credentials",
			body: map[string]string{
				"username": "admin",
				"password": "admin",
			},
			expectedStatus: http.StatusOK,
			expectToken:    true,
		},
		{
			name: "invalid credentials",
			body: map[string]string{
				"username": "baduser",
				"password": "badpass",
			},
			expectedStatus: http.StatusUnauthorized,
			expectToken:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonValue, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonValue))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedStatus, resp.Code)

			var data map[string]string
			json.Unmarshal(resp.Body.Bytes(), &data)

			if tt.expectToken {
				assert.Contains(t, data, "token")
			} else {
				assert.NotContains(t, data, "token")
			}
		})
	}
}

func TestProtectedRoute(t *testing.T) {
	router := setupRouter()

	// Step 1: get valid token
	token := ""
	{
		body := map[string]string{"username": "admin", "password": "admin"}
		jsonValue, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		var data map[string]string
		json.Unmarshal(resp.Body.Bytes(), &data)
		token = data["token"]
	}

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "missing token",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "valid token",
			authHeader:     "Bearer " + token,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/secret/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedStatus, resp.Code)
		})
	}
}
