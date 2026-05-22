package model

import "time"

type Employee struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	DepartmentID uint       `gorm:"not null" json:"department_id"`
	FullName     string     `gorm:"type:varchar(200);not null" json:"full_name"`
	Position     string     `gorm:"type:varchar(200);not null" json:"position"`
	HiredAt      *time.Time `gorm:"type:date" json:"hired_at"`
	CreatedAt    time.Time  `json:"created_at"`
}
