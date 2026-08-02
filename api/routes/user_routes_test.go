package routes

import (
	"learnlang-api/config"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSetupUserRoutesRegistersExpectedEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api")
	SetupUserRoutes(api, &config.Config{JWT: config.JWTConfig{Secret: "secret"}}, nil, &Services{})
	routes := router.Routes()
	if len(routes) != 29 {
		t.Fatalf("registered %d routes, want 29", len(routes))
	}
	wanted := map[string]bool{
		"POST /api/user/auth/login":              false,
		"POST /api/user/chat":                    false,
		"GET /api/user/profile":                  false,
		"GET /api/user/ws/chat":                  false,
		"POST /api/user/vocabularies/:id/import": false,
		"PUT /api/user/vocabularies/:id/default": false,
		"GET /api/user/usage/events":             false,
		"GET /api/user/usage/summary":            false,
	}
	for _, route := range routes {
		key := route.Method + " " + route.Path
		if _, ok := wanted[key]; ok {
			wanted[key] = true
		}
	}
	for route, found := range wanted {
		if !found {
			t.Errorf("route %s was not registered", route)
		}
	}
}
