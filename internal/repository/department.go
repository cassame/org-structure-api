package repository

import (
	"apiOrganizationStructure/internal/model"
	"context"
	"errors"

	"gorm.io/gorm"
)

type DepartmentRepository struct {
	db *gorm.DB
}

func NewDepartmentRepository(db *gorm.DB) *DepartmentRepository {
	return &DepartmentRepository{db: db}
}

func (r *DepartmentRepository) Create(ctx context.Context, dept *model.Department) error {
	return r.db.WithContext(ctx).Create(dept).Error
}

func (r *DepartmentRepository) GetByID(ctx context.Context, id uint) (*model.Department, error) {
	var dept model.Department
	err := r.db.WithContext(ctx).First(&dept, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &dept, err
}

func (r *DepartmentRepository) ReassignEmployees(ctx context.Context, fromDeptID, toDeptID uint) error {
	return r.db.WithContext(ctx).
		Model(&model.Employee{}).
		Where("department_id = ?", fromDeptID).
		Update("department_id", toDeptID).Error
}

func (r *DepartmentRepository) Update(ctx context.Context, dept *model.Department) error {
	return r.db.WithContext(ctx).Save(dept).Error
}

func (r *DepartmentRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Department{}, id).Error
}

func (r *DepartmentRepository) GetAll(ctx context.Context) ([]model.Department, error) {
	var depts []model.Department
	err := r.db.WithContext(ctx).Find(&depts).Error
	return depts, err
}

func (r *DepartmentRepository) GetAllWithEmployees(ctx context.Context) ([]model.Department, error) {
	var depts []model.Department
	err := r.db.WithContext(ctx).Preload("Employees").Find(&depts).Error
	return depts, err
}
