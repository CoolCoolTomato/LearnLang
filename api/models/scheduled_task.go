package models

import "time"

type ScheduledTask struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       int64     `gorm:"default:0" json:"user_id"`
	FunctionName string    `gorm:"size:64;not null" json:"function_name"`
	Args         string    `gorm:"type:text" json:"args"`
	ScheduledAt  time.Time `gorm:"not null" json:"scheduled_at"`
	Status       string    `gorm:"size:16;default:pending" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

func (ScheduledTask) TableName() string {
	return "scheduled_tasks"
}
