package models

import (
	"time"
	"gorm.io/gorm"
)

type Pomodoro struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	IsCompleted bool      `json:"is_completed"`
	Duration    int       `json:"duration"` // in minutes
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}