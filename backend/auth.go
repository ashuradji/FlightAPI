package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("your_secret_key")

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func LoginHandler(c *gin.Context) {
	var creds Credentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Dummy user check
	if creds.Username != "admin" || creds.Password != "admin" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate token
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &jwt.RegisteredClaims{
		Subject:   creds.Username,
		ExpiresAt: jwt.NewNumericDate(expirationTime),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}
