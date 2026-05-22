package repository

import (
	"apiOrganizationStructure/internal/model"
	"context"
	"errors"

	"gorm.io/gorm"
)

type EmployeeRepository struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) *EmployeeRepository {
	return &EmployeeRepository{db: db}
}

func (r *EmployeeRepository) Create(ctx context.Context, emp *model.Employee) error {
	return r.db.WithContext(ctx).Create(emp).Error
}

func (r *EmployeeRepository) GetByID(ctx context.Context, id uint) (*model.Employee, error) {
	var emp model.Employee
	err := r.db.WithContext(ctx).First(&emp, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &emp, err
}

func (r *EmployeeRepository) Update(ctx context.Context, emp *model.Employee) error {
	return r.db.WithContext(ctx).Save(emp).Error
}

func (r *EmployeeRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Employee{}, id).Error
}
