package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"learnlang-api/database"
	"learnlang-api/models"
	"strings"
)

type UserProfileSummaryTool struct {
	UserID int64
}

func (t UserProfileSummaryTool) Name() string {
	return "update_user_profile_summary"
}

func (t UserProfileSummaryTool) Description() string {
	return `Update the user's stable profile summary. Call only when durable profile information should change. Input must be JSON: {"summary":"updated stable user profile summary"}.`
}

func (t UserProfileSummaryTool) Call(ctx context.Context, input string) (string, error) {
	var args struct {
		Summary string `json:"summary"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("parse update_user_profile_summary input: %w", err)
	}

	args.Summary = strings.TrimSpace(args.Summary)
	if args.Summary == "" {
		return "", fmt.Errorf("summary is required")
	}

	var profile models.UserProfileSummary
	if err := database.DB.WithContext(ctx).Where("user_id = ?", t.UserID).First(&profile).Error; err != nil {
		profile = models.UserProfileSummary{UserID: t.UserID}
	}

	profile.Summary = args.Summary
	if err := database.DB.WithContext(ctx).Save(&profile).Error; err != nil {
		return "", err
	}

	return marshalToolResult(map[string]any{
		"status":  "updated",
		"summary": profile.Summary,
	})
}
