package model

import "time"

type Department struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	Name      string    `gorm:"type:varchar(200);not null;index:idx_parent_name,unique" json:"name"`
	ParentID  *uint     `gorm:"index:idx_parent_name,unique" json:"parent_id"`
	CreatedAt time.Time `json:"created_at"`

	Parent    *Department  `gorm:"foreignkey:ParentID;constraint:OnDelete:CASCADE;" json:"-"`
	Employees []Employee   `gorm:"foreignKey:DepartmentID;constraint:OnDelete:CASCADE;" json:"employees,omitempty"`
	Children  []Department `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}
