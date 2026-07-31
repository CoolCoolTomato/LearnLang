package dev

import (
	"errors"
	"learnlang-api/services"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func devContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(method, target, strings.NewReader(body))
	context.Request.Header.Set("Content-Type", "application/json")
	return context, recorder
}

func TestControllerIDAndErrorMapping(t *testing.T) {
	controller := NewController(nil, nil)
	context, _ := devContext(http.MethodGet, "/", "")
	context.Params = gin.Params{{Key: "id", Value: "8"}}
	if id, ok := controller.id(context); !ok || id != 8 {
		t.Fatalf("id() = %d, %v", id, ok)
	}
	context, recorder := devContext(http.MethodGet, "/", "")
	context.Params = gin.Params{{Key: "id", Value: "bad"}}
	if _, ok := controller.id(context); ok || recorder.Code != http.StatusBadRequest {
		t.Fatalf("invalid id status = %d", recorder.Code)
	}
	for _, tt := range []struct {
		err  error
		code int
	}{
		{gorm.ErrRecordNotFound, http.StatusNotFound},
		{errors.New("ids are required"), http.StatusBadRequest},
		{errors.New("query is required"), http.StatusBadRequest},
		{errors.New("at least one writable field is required"), http.StatusBadRequest},
		{errors.New(`unknown developer resource "bad"`), http.StatusBadRequest},
		{errors.New("failure"), http.StatusInternalServerError},
	} {
		context, recorder = devContext(http.MethodGet, "/", "")
		controller.writeError(context, tt.err)
		if recorder.Code != tt.code {
			t.Errorf("writeError(%v) = %d, want %d", tt.err, recorder.Code, tt.code)
		}
	}
}

func TestDeveloperControllerValidation(t *testing.T) {
	controller := NewController(services.NewDeveloperDataService(nil), nil)
	context, recorder := devContext(http.MethodGet, "/", "")
	controller.Dashboard(context)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("Dashboard() = %d %s", recorder.Code, recorder.Body.String())
	}
	context, recorder = devContext(http.MethodPost, "/", `{}`)
	controller.SearchConversationArchives(context)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("Search() unauthenticated = %d", recorder.Code)
	}
	context, recorder = devContext(http.MethodPost, "/", `{}`)
	context.Set("user_id", int64(1))
	controller.SearchConversationArchives(context)
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("Search() nil service = %d %s", recorder.Code, recorder.Body.String())
	}
	for _, call := range []func(*gin.Context){controller.Create, controller.DeleteMany} {
		context, recorder = devContext(http.MethodPost, "/", `{`)
		call(context)
		if recorder.Code != http.StatusBadRequest {
			t.Errorf("invalid JSON status = %d", recorder.Code)
		}
	}
}
