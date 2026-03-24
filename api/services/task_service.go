package services

import (
	"context"
	"learnlang-api/database"
	"learnlang-api/models"
	"time"
)

type ScheduledTaskService struct {
	handlers map[string]TaskHandler
}

type TaskHandler func(args string) error

func NewScheduledTaskService() *ScheduledTaskService {
	return &ScheduledTaskService{
		handlers: make(map[string]TaskHandler),
	}
}

func (s *ScheduledTaskService) RegisterHandler(functionName string, handler TaskHandler) {
	s.handlers[functionName] = handler
}

func (s *ScheduledTaskService) CreateTask(userID int64, functionName string, args string, scheduledAt time.Time) (*models.ScheduledTask, error) {
	task := models.ScheduledTask{
		UserID:       userID,
		FunctionName: functionName,
		Args:         args,
		ScheduledAt:  scheduledAt,
		Status:       "pending",
	}
	if err := database.DB.Create(&task).Error; err != nil {
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

func (s *ScheduledTaskService) UpdateTask(task *models.ScheduledTask, updates map[string]interface{}) error {
	return database.DB.Model(task).Updates(updates).Error
}

func (s *ScheduledTaskService) DeleteTask(id int64) error {
	return database.DB.Delete(&models.ScheduledTask{}, id).Error
}

func (s *ScheduledTaskService) CancelUserPendingTasks(userID int64, functionName string) error {
	return database.DB.Model(&models.ScheduledTask{}).
		Where("user_id = ? AND function_name = ? AND status = ?", userID, functionName, "pending").
		Updates(map[string]interface{}{
			"status": "cancelled",
		}).Error
}
