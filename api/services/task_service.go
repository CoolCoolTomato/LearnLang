package services

import (
	"context"
	"fmt"
	"learnlang-api/database"
	"learnlang-api/models"
	"time"
)

type ScheduledTaskService struct {
	handlers map[string]TaskHandler
}

type TaskHandler func(args string) error

const (
	ScheduledTaskFilterAll        = "all"
	ScheduledTaskFilterUnfinished = "unfinished"
	ScheduledTaskFilterCompleted  = "completed"
)

type ScheduledTaskPage struct {
	Tasks      []models.ScheduledTask
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
	HasNext    bool
}

func NewScheduledTaskService() *ScheduledTaskService {
	return &ScheduledTaskService{
		handlers: make(map[string]TaskHandler),
	}
}

func (s *ScheduledTaskService) RegisterHandler(functionName string, handler TaskHandler) {
	s.handlers[functionName] = handler
}

func (s *ScheduledTaskService) CreateTask(ctx context.Context, userID int64, functionName string, args string, scheduledAt time.Time) (*models.ScheduledTask, error) {
	task := models.ScheduledTask{
		UserID:       userID,
		FunctionName: functionName,
		Args:         args,
		ScheduledAt:  scheduledAt.UTC(),
		Status:       "pending",
	}
	if err := database.DB.WithContext(ctx).Create(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *ScheduledTaskService) StartScheduler(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.processPendingTasks()
		}
	}
}

func (s *ScheduledTaskService) processPendingTasks() {
	var tasks []models.ScheduledTask
	now := time.Now().UTC()

	database.DB.Where("status = ? AND scheduled_at <= ?", "pending", now).Find(&tasks)

	for _, task := range tasks {
		s.executeTask(&task)
	}
}

func (s *ScheduledTaskService) executeTask(task *models.ScheduledTask) {
	handler, ok := s.handlers[task.FunctionName]
	if !ok {
		task.Status = "failed"
		database.DB.Save(task)
		return
	}

	if err := handler(task.Args); err != nil {
		task.Status = "failed"
	} else {
		task.Status = "completed"
	}
	database.DB.Save(task)
}

func (s *ScheduledTaskService) GetTask(id int64) (*models.ScheduledTask, error) {
	var task models.ScheduledTask
	if err := database.DB.First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *ScheduledTaskService) ListTasks(userID *int64, status *string, page int, size int) ([]models.ScheduledTask, error) {
	var tasks []models.ScheduledTask
	query := database.DB.Model(&models.ScheduledTask{})

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	offset := (page - 1) * size
	if err := query.Order("scheduled_at DESC").Limit(size).Offset(offset).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *ScheduledTaskService) ListUserTasks(ctx context.Context, userID int64, filter string, page, pageSize int) (*ScheduledTaskPage, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("user_id must be positive")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}

	query := database.DB.WithContext(ctx).Model(&models.ScheduledTask{}).Where("user_id = ?", userID)
	switch filter {
	case "", ScheduledTaskFilterAll:
		filter = ScheduledTaskFilterAll
	case ScheduledTaskFilterUnfinished:
		query = query.Where("status <> ?", "completed")
	case ScheduledTaskFilterCompleted:
		query = query.Where("status = ?", "completed")
	default:
		return nil, fmt.Errorf("unsupported scheduled task filter %q", filter)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	tasks := make([]models.ScheduledTask, 0)
	if err := query.Order("scheduled_at DESC, id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return &ScheduledTaskPage{
		Tasks: tasks, Total: total, Page: page, PageSize: pageSize,
		TotalPages: totalPages, HasNext: page < totalPages,
	}, nil
}

func (s *ScheduledTaskService) UpdateTask(task *models.ScheduledTask, updates map[string]interface{}) error {
	return database.DB.Model(task).Updates(updates).Error
}

func (s *ScheduledTaskService) DeleteTask(id int64) error {
	return database.DB.Delete(&models.ScheduledTask{}, id).Error
}
