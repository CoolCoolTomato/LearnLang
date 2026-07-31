package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestPasswordHashAndCheck(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if hash == "correct horse battery staple" {
		t.Fatal("HashPassword() returned the plaintext password")
	}
	if !CheckPassword("correct horse battery staple", hash) {
		t.Fatal("CheckPassword() rejected the correct password")
	}
	if CheckPassword("wrong", hash) {
		t.Fatal("CheckPassword() accepted an incorrect password")
	}
	if CheckPassword("password", "not-a-bcrypt-hash") {
		t.Fatal("CheckPassword() accepted an invalid hash")
	}
}

func TestGenerateAndParseToken(t *testing.T) {
	before := time.Now().UTC()
	token, err := GenerateToken(42, "admin", "secret")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	claims, err := ParseToken(token, "secret")
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if claims.UserID != 42 || claims.Role != "admin" {
		t.Fatalf("ParseToken() claims = %#v", claims)
	}
	if claims.IssuedAt == nil || claims.ExpiresAt == nil {
		t.Fatal("token timestamps are missing")
	}
	if claims.IssuedAt.Time.Before(before.Add(-time.Second)) {
		t.Fatalf("IssuedAt = %v, want near %v", claims.IssuedAt.Time, before)
	}
	if got := claims.ExpiresAt.Sub(claims.IssuedAt.Time); got != 24*time.Hour {
		t.Fatalf("token lifetime = %v, want 24h", got)
	}
}

func TestParseTokenRejectsInvalidTokens(t *testing.T) {
	valid, err := GenerateToken(1, "user", "right-secret")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseToken(valid, "wrong-secret"); err == nil {
		t.Fatal("ParseToken() accepted a token signed with another secret")
	}
	if _, err := ParseToken("not-a-token", "secret"); err == nil {
		t.Fatal("ParseToken() accepted malformed input")
	}

	expiredClaims := Claims{
		UserID: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	}
	expired, err := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims).SignedString([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseToken(expired, "secret"); err == nil {
		t.Fatal("ParseToken() accepted an expired token")
	}
}
