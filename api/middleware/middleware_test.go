package middleware

import (
	"learnlang-api/database"
	"learnlang-api/models"
	"learnlang-api/utils"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func performMiddlewareRequest(handler gin.HandlerFunc, request *http.Request) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	router := gin.New()
	router.Use(handler)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"user_id": c.GetInt64("user_id"), "role": c.GetString("role")})
	})
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestAuthMiddleware(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	manager := utils.NewTokenManager(client)
	token, err := utils.GenerateToken(7, "admin", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if err := manager.SaveToken(7, token, 0); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := performMiddlewareRequest(AuthMiddleware("secret", manager), request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"user_id":7`) || !strings.Contains(response.Body.String(), `"role":"admin"`) {
		t.Fatalf("valid response = %d %s", response.Code, response.Body.String())
	}

	tests := []struct {
		name   string
		header string
		code   int
		body   string
	}{
		{"missing", "", http.StatusUnauthorized, "header required"},
		{"format", "Basic abc", http.StatusUnauthorized, "format"},
		{"invalid", "Bearer invalid", http.StatusUnauthorized, "Invalid token"},
		{"revoked", "Bearer " + token, http.StatusUnauthorized, "revoked"},
	}
	server.Del("token:7")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.header != "" {
				request.Header.Set("Authorization", tt.header)
			}
			response := performMiddlewareRequest(AuthMiddleware("secret", manager), request)
			if response.Code != tt.code || !strings.Contains(response.Body.String(), tt.body) {
				t.Fatalf("response = %d %s", response.Code, response.Body.String())
			}
		})
	}

	server.Close()
	request = httptest.NewRequest(http.MethodGet, "/test", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response = performMiddlewareRequest(AuthMiddleware("secret", manager), request)
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("Redis error response = %d %s", response.Code, response.Body.String())
	}
}

func TestWebSocketAuthMiddlewareAcceptsQueryToken(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	manager := utils.NewTokenManager(client)
	token, _ := utils.GenerateToken(4, "user", "secret")
	_ = manager.SaveToken(4, token, 0)

	request := httptest.NewRequest(http.MethodGet, "/test?token="+token, nil)
	response := performMiddlewareRequest(WebSocketAuthMiddleware("secret", manager), request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"user_id":4`) {
		t.Fatalf("query token response = %d %s", response.Code, response.Body.String())
	}
	response = performMiddlewareRequest(WebSocketAuthMiddleware("secret", manager), httptest.NewRequest(http.MethodGet, "/test", nil))
	if response.Code != http.StatusUnauthorized || !strings.Contains(response.Body.String(), "Token required") {
		t.Fatalf("missing token response = %d %s", response.Code, response.Body.String())
	}
}

func TestDeveloperMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, role := range []string{"developer", "admin"} {
		recorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(recorder)
		context.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		context.Set("role", role)
		DeveloperMiddleware("")(context)
		if context.IsAborted() {
			t.Errorf("role %q was rejected", role)
		}
	}

	db, err := gorm.Open(sqlite.Open("file:middleware_developer?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatal(err)
	}
	user := models.User{Username: "initial", PasswordHash: "hash"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	previous := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = previous })

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	context.Set("role", "user")
	context.Set("user_id", user.ID)
	DeveloperMiddleware("initial")(context)
	if context.IsAborted() {
		t.Fatalf("initial user rejected: %s", recorder.Body.String())
	}

	recorder = httptest.NewRecorder()
	context, _ = gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	context.Set("role", "user")
	DeveloperMiddleware("initial")(context)
	if !context.IsAborted() || recorder.Code != http.StatusForbidden {
		t.Fatalf("ordinary user response = %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestCORS(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	response := performMiddlewareRequest(CORS(), request)
	if response.Code != http.StatusOK || response.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("GET CORS response = %d %#v", response.Code, response.Header())
	}
	request = httptest.NewRequest(http.MethodOptions, "/test", nil)
	response = performMiddlewareRequest(CORS(), request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("OPTIONS response = %d", response.Code)
	}
}
