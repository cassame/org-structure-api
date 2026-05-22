package service

import (
	"apiOrganizationStructure/internal/model"
	"apiOrganizationStructure/internal/repository"
	"context"
	"time"
)

type EmployeeRepository interface {
	Create(ctx context.Context, emp *model.Employee) error
	GetByID(ctx context.Context, id uint) (*model.Employee, error)
	Update(ctx context.Context, emp *model.Employee) error
	Delete(ctx context.Context, id uint) error
}
type EmployeeService struct {
	empRepo  EmployeeRepository
	deptRepo DepartmentRepository
}

func NewEmployeeService(empRepo *repository.EmployeeRepository, deptRepo *repository.DepartmentRepository) *EmployeeService {
	return &EmployeeService{
		empRepo:  empRepo,
		deptRepo: deptRepo,
	}
}

func (s *EmployeeService) Create(ctx context.Context, deptID uint, name, position string, hiredAt *time.Time) (*model.Employee, error) {
	dept, err := s.deptRepo.GetByID(ctx, deptID)
	if err != nil {
		return nil, err
	}
	if dept == nil {
		return nil, ErrParentNotFound
	}

	emp := &model.Employee{
		DepartmentID: dept.ID,
		FullName:     name,
		Position:     position,
		HiredAt:      hiredAt,
	}

	if err := s.empRepo.Create(ctx, emp); err != nil {
		return nil, err
	}
	return emp, nil
}

func (s *EmployeeService) Update(ctx context.Context, id uint, deptID uint, name, position string, hiredAt *time.Time) (*model.Employee, error) {
	emp, err := s.empRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if emp == nil {
		return nil, ErrEmployeeNotFound
	}

	if emp.DepartmentID != deptID {
		dept, err := s.deptRepo.GetByID(ctx, deptID)
		if err != nil {
			return nil, err
		}
		if dept == nil {
			return nil, ErrParentNotFound
		}
	}

	emp.DepartmentID = deptID
	emp.FullName = name
	emp.Position = position
	emp.HiredAt = hiredAt

	if err := s.empRepo.Update(ctx, emp); err != nil {
		return nil, err
	}

	return emp, nil
}

func (s *EmployeeService) Delete(ctx context.Context, id uint) error {
	emp, err := s.empRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if emp == nil {
		return ErrEmployeeNotFound
	}
	return s.empRepo.Delete(ctx, id)
}
