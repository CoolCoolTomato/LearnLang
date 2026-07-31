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
	if len(routes) != 26 {
		t.Fatalf("registered %d routes, want 26", len(routes))
	}
	wanted := map[string]bool{
		"POST /api/user/auth/login":              false,
		"POST /api/user/chat":                    false,
		"GET /api/user/profile":                  false,
		"GET /api/user/ws/chat":                  false,
		"POST /api/user/vocabularies/:id/import": false,
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
