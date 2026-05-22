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
	ReassignEmployees(ctx context.Context, fromDeptID, toDeptID uint) error
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

func (s *DepartmentService) GetByID(ctx context.Context, id uint, depth int, withEmployees bool) (*model.Department, error) {
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
		allDepts[i].Children = make([]model.Department, 0)
		if allDepts[i].Employees == nil {
			allDepts[i].Employees = make([]model.Employee, 0)
		}
		idToDept[allDepts[i].ID] = &allDepts[i]
	}

	dept, exists := idToDept[id]
	if !exists {
		return nil, ErrDepartmentNotFound
	}

	result := buildSubtree(dept, idToDept, depth)
	return &result, nil
}

func buildSubtree(dept *model.Department, idToDept map[uint]*model.Department, depth int) model.Department {
	node := *dept
	node.Children = make([]model.Department, 0)

	if depth <= 0 {
		return node
	}

	for _, candidate := range idToDept {
		if candidate.ParentID != nil && *candidate.ParentID == node.ID {
			child := buildSubtree(candidate, idToDept, depth-1)
			node.Children = append(node.Children, child)
		}
	}
	return node
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

func (s *DepartmentService) Delete(ctx context.Context, id uint, mode string, reassignTo *uint) error {
	dept, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if dept == nil {
		return ErrDepartmentNotFound
	}

	switch mode {
	case "cascade":
		return s.repo.Delete(ctx, id)

	case "reassign":
		if reassignTo == nil {
			return ErrInvalidDeleteMode
		}
		target, err := s.repo.GetByID(ctx, *reassignTo)
		if err != nil {
			return err
		}
		if target == nil {
			return ErrReassignTargetNotFound
		}
		if err := s.repo.ReassignEmployees(ctx, id, *reassignTo); err != nil {
			return err
		}
		return s.repo.Delete(ctx, id)

	default:
		return ErrInvalidDeleteMode
	}
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
				parent.Children = append(parent.Children, allDepts[i])
			}
		}
	}

	var finalTree []model.Department
	for _, dept := range idToDept {
		if dept.ParentID == nil {
			finalTree = append(finalTree, *dept)
		}
	}

	return finalTree, nil
}
