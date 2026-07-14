package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"learnlang-api/database"
	"learnlang-api/models"
	"strings"

	"gorm.io/gorm"
)

type UserProfileSummaryTool struct {
	UserID int64
}

func (t UserProfileSummaryTool) Name() string {
	return "update_user_profile_summary"
}

func (t UserProfileSummaryTool) Description() string {
	return `Update the user's concise, evidence-based personal portrait. The full profile should contain only applicable sections: Basic facts; Interests and preferences; Goals and life context; Interaction impression. Explicit facts such as name, birthday, occupation, relationships, goals, favorites, and strong preferences are saved after one clear self-report. Interaction impressions require explicit self-description or consistent evidence from at least two separate interactions and must be phrased as tendencies, not facts. Input must be JSON: {"summary":"complete updated user profile"}. Preserve unchanged profile items, replace only explicitly corrected facts, deduplicate content, and never add conversation logs, temporary states, task details, or unsupported inferences.`
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
	err := database.DB.WithContext(ctx).Where("user_id = ?", t.UserID).First(&profile).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		profile = models.UserProfileSummary{UserID: t.UserID}
	}
	if profile.Summary == args.Summary {
		return marshalToolResult(map[string]any{
			"status":  "unchanged",
			"summary": profile.Summary,
		})
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
