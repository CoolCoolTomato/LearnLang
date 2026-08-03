package tools

import (
	"context"
	"encoding/json"
	"learnlang-api/models"
	"learnlang-api/services"
	"testing"
	"time"
)

type fakeScheduledTaskLister struct {
	userID   int64
	filter   string
	page     int
	pageSize int
}

func (f *fakeScheduledTaskLister) ListScheduledTasks(_ context.Context, userID int64, filter string, page, pageSize int) (*services.ScheduledTaskPage, error) {
	f.userID, f.filter, f.page, f.pageSize = userID, filter, page, pageSize
	args, _ := json.Marshal(services.SendMessageArgs{UserID: userID, Message: "Review words", Translation: "复习单词"})
	return &services.ScheduledTaskPage{
		Tasks: []models.ScheduledTask{{
			ID: 8, FunctionName: "send_message", Args: string(args), Status: "pending",
			ScheduledAt: time.Date(2026, 8, 3, 2, 0, 0, 0, time.UTC),
			CreatedAt:   time.Date(2026, 8, 2, 1, 0, 0, 0, time.UTC),
		}},
		Total: 11, Page: page, PageSize: pageSize, TotalPages: 2, HasNext: true,
	}, nil
}

func TestScheduledTaskQueryToolReturnsPaginatedUserTasks(t *testing.T) {
	lister := &fakeScheduledTaskLister{}
	tool := ScheduledTaskQueryTool{UserID: 7, Timezone: "Asia/Shanghai", Lister: lister}
	output, err := tool.Call(context.Background(), `{"status":"unfinished","page":1,"page_size":10}`)
	if err != nil {
		t.Fatal(err)
	}
	if lister.userID != 7 || lister.filter != services.ScheduledTaskFilterUnfinished || lister.page != 1 || lister.pageSize != 10 {
		t.Fatalf("query = user %d filter %q page %d size %d", lister.userID, lister.filter, lister.page, lister.pageSize)
	}
	var result struct {
		Status     string `json:"status"`
		Timezone   string `json:"timezone"`
		Total      int64  `json:"total"`
		TotalPages int    `json:"total_pages"`
		HasNext    bool   `json:"has_next"`
		Tasks      []struct {
			Status       string `json:"status"`
			ScheduledAt  string `json:"scheduled_at"`
			Message      string `json:"message"`
			Translation  string `json:"translation"`
			DetailsReady bool   `json:"details_available"`
		} `json:"tasks"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != "ok" || result.Timezone != "Asia/Shanghai" || result.Total != 11 || result.TotalPages != 2 || !result.HasNext || len(result.Tasks) != 1 {
		t.Fatalf("result = %#v", result)
	}
	item := result.Tasks[0]
	if item.Status != "pending" || item.ScheduledAt != "2026-08-03T10:00:00+08:00" || item.Message != "Review words" || item.Translation != "复习单词" || !item.DetailsReady {
		t.Fatalf("task = %#v", item)
	}
}

func TestScheduledTaskQueryToolDefaultsAndValidation(t *testing.T) {
	lister := &fakeScheduledTaskLister{}
	tool := ScheduledTaskQueryTool{UserID: 7, Timezone: "invalid", Lister: lister}
	output, err := tool.Call(context.Background(), `{}`)
	if err != nil || lister.filter != services.ScheduledTaskFilterAll || lister.page != 1 || lister.pageSize != defaultScheduledTaskPageSize {
		t.Fatalf("default query output = %q, error = %v", output, err)
	}
	if !json.Valid([]byte(output)) {
		t.Fatalf("output is not JSON: %s", output)
	}
	invalid, err := tool.Call(context.Background(), `{"status":"pending"}`)
	if err != nil || !json.Valid([]byte(invalid)) || !containsJSONText(invalid, "invalid_request") {
		t.Fatalf("invalid status output = %q, error = %v", invalid, err)
	}
}

func containsJSONText(value, expected string) bool {
	var result map[string]any
	if json.Unmarshal([]byte(value), &result) != nil {
		return false
	}
	return result["status"] == expected
}
