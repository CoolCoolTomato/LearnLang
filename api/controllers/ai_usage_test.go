package controllers

import (
	"encoding/json"
	"learnlang-api/models"
	"net/http"
	"testing"
	"time"
)

func TestVisibleAIUsageEventContainsOnlyUserFields(t *testing.T) {
	event := visibleAIUsageEvent(models.AIUsageEvent{
		ID: 9, UserID: 7, Operation: models.AIOperationChat, Model: "model",
		Usage: 42, Unit: models.AIUsageUnitTokens, Status: models.AIUsageStatusSucceeded,
		CreatedAt: time.Date(2026, 8, 2, 10, 0, 0, 0, time.FixedZone("CST", 8*60*60)),
	})
	encoded, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	wanted := []string{"operation", "model", "usage", "unit", "status", "created_at"}
	if len(fields) != len(wanted) {
		t.Fatalf("visible fields = %v", fields)
	}
	for _, key := range wanted {
		if _, ok := fields[key]; !ok {
			t.Errorf("missing visible field %q", key)
		}
	}
}

func TestAIUsageControllerRequiresAuthenticationAndValidQuery(t *testing.T) {
	controller := NewAIUsageController(nil)
	context, recorder := controllerContext(http.MethodGet, "/", "")
	controller.List(context)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d", recorder.Code)
	}
	context, recorder = controllerContext(http.MethodGet, "/?page=bad", "")
	context.Set("user_id", int64(1))
	controller.List(context)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("invalid query status = %d", recorder.Code)
	}
}
