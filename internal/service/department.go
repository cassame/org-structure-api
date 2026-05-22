package service

import (
	"apiOrganizationStructure/internal/model"
	"context"
	"errors"
)

type DepartmentRepository interface {
	Create(ctx context.Context, dept *model.Department) error
	GetByID(ctx context.Context, id uint) (*model.Department, error)
	Update(ctx context.Context, dept *model.Department) error
	Delete(ctx context.Context, id uint) error

	GetAll(ctx context.Context) ([]model.Department, error)
	GetAllWithEmployees(ctx context.Context) ([]model.Department, error)
}

type DepartmentService struct {
	repo DepartmentRepository
}

func NewDepartmentService(repo DepartmentRepository) *DepartmentService {
	return &DepartmentService{repo: repo}
}

func (s *DepartmentService) Create(ctx context.Context, name string, parentID *uint) (*model.Department, error) {
	if parentID != nil {
		parent, err := s.repo.GetByID(ctx, *parentID)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			return nil, ErrParentNotFound
		}
	}
	dept := &model.Department{
		Name:     name,
		ParentID: parentID,
	}
	if err := s.repo.Create(ctx, dept); err != nil {
		return nil, err
	}
	return dept, nil
}

func (s *DepartmentService) Update(ctx context.Context, id uint, newName string, newParentID *uint) (*model.Department, error) {
	dept, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if dept == nil {
		return nil, ErrDepartmentNotFound
	}

	if newParentID != nil {
		if *newParentID == id {
			return nil, errors.New("department cannot be its own parent")
		}

		isCyclic, err := s.checkCycle(ctx, id, *newParentID)
		if err != nil {
			return nil, err
		}
		if isCyclic {
			return nil, ErrCycleDetected
		}
	}
	dept.Name = newName
	dept.ParentID = newParentID

	if err := s.repo.Update(ctx, dept); err != nil {
		return nil, err
	}

	return dept, nil
}

func (s *DepartmentService) checkCycle(ctx context.Context, currentID uint, targetParentID uint) (bool, error) {
	allDepts, err := s.repo.GetAll(ctx)
	if err != nil {
		return false, err
	}
	parentToChildren := make(map[uint][]uint)
	for _, d := range allDepts {
		if d.ParentID != nil {
			parentToChildren[*d.ParentID] = append(parentToChildren[*d.ParentID], d.ID)
		}
	}

	queue := []uint{currentID}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		for _, childID := range parentToChildren[curr] {
			if childID == targetParentID {
				return true, nil
			}
			queue = append(queue, childID)
		}
	}
	return false, nil
}

func (s *DepartmentService) Delete(ctx context.Context, id uint) error {
	dept, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if dept == nil {
		return ErrDepartmentNotFound
	}
	return s.repo.Delete(ctx, id)
}

func (s *DepartmentService) GetTree(ctx context.Context, withEmployees bool) ([]model.Department, error) {
	var allDepts []model.Department
	var err error

	if withEmployees {
		allDepts, err = s.repo.GetAllWithEmployees(ctx)
	} else {
		allDepts, err = s.repo.GetAll(ctx)
	}
	if err != nil {
		return nil, err
	}

	idToDept := make(map[uint]*model.Department)

	for i := range allDepts {
		if allDepts[i].Children == nil {
			allDepts[i].Children = make([]model.Department, 0)
		}
		if allDepts[i].Employees == nil {
			allDepts[i].Employees = make([]model.Employee, 0)
		}
		idToDept[allDepts[i].ID] = &allDepts[i]
	}

	for i := range allDepts {
		dept := idToDept[allDepts[i].ID]
		if dept.ParentID != nil {
			parent, exists := idToDept[*dept.ParentID]
			if exists {
				parent.Children = append(parent.Children, *dept)
			}
		}
	}

	var finalTree []model.Department
	for i := range allDepts {
		if allDepts[i].ParentID == nil {
			finalTree = append(finalTree, *idToDept[allDepts[i].ID])
		}
	}

	return finalTree, nil
}
