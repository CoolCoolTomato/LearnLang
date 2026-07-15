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
	return `Update or clean the user's concise, evidence-based personal portrait. The full profile should contain only applicable sections: Basic facts; Interests and preferences; Goals and life context; Interaction impression. Explicit facts such as name, birthday, occupation, relationships, goals, favorites, and strong preferences are saved after one clear self-report. Convert relative dates to stable dates using the current user time; never save words such as today or tomorrow. Interaction impressions require explicit self-description or consistent evidence from at least two separate interactions and must be phrased as tendencies, not facts. Input must be JSON: {"summary":"complete updated user profile"}. Preserve valid unchanged facts, correct contradictions, deduplicate content, and remove conversation logs, greetings, queried topics, missing-information notes, temporary states, task details, and unsupported inferences from legacy profile content.`
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
	if reason := validateUserProfileSummary(args.Summary); reason != "" {
		return marshalToolResult(map[string]any{
			"status": "rejected",
			"error":  reason,
		})
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

func validateUserProfileSummary(summary string) string {
	lower := strings.ToLower(summary)
	relativeDates := []string{"今天", "明天", "昨天", "today", "tomorrow", "yesterday"}
	for _, phrase := range relativeDates {
		if strings.Contains(lower, phrase) {
			return fmt.Sprintf("profile contains relative date %q; convert it to an absolute date using the current user time", phrase)
		}
	}

	conversationLogPhrases := []string{
		"用户询问过",
		"用户提到过",
		"用户讨论过",
		"用户进行日常问候",
		"进行了日常问候",
		"具体类型未明确",
		"尚不清楚",
		"asked about",
		"discussed",
		"not specified",
	}
	for _, phrase := range conversationLogPhrases {
		if strings.Contains(lower, phrase) {
			return fmt.Sprintf("profile contains conversation-log or missing-information phrase %q; remove it or convert an explicit durable self-report into a person fact", phrase)
		}
	}
	return ""
}
