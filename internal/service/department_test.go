package service

import (
	"apiOrganizationStructure/internal/model"
	"context"
	"errors"
	"testing"
)

type mockDepartmentRepository struct {
	getByIDFunc func(ctx context.Context, id uint) (*model.Department, error)
	updateFunc  func(ctx context.Context, dept *model.Department) error
}

func (m *mockDepartmentRepository) Create(ctx context.Context, dept *model.Department) error {
	return nil
}

func (m *mockDepartmentRepository) GetByID(ctx context.Context, id uint) (*model.Department, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockDepartmentRepository) Update(ctx context.Context, dept *model.Department) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, dept)
	}
	return nil
}

func (m *mockDepartmentRepository) Delete(ctx context.Context, id uint) error {
	return nil
}

func (m *mockDepartmentRepository) GetAll(ctx context.Context) ([]model.Department, error) {
	return nil, nil
}

func (m *mockDepartmentRepository) GetAllWithEmployees(ctx context.Context) ([]model.Department, error) {
	return nil, nil
}

func TestDepartmentService_Update(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		deptID        uint
		newName       string
		newParentID   *uint
		mockGetByID   func(ctx context.Context, id uint) (*model.Department, error)
		expectedError error
	}{
		{
			name:        "Успешное обновление без смены родителя",
			deptID:      1,
			newName:     "New Name",
			newParentID: nil,
			mockGetByID: func(ctx context.Context, id uint) (*model.Department, error) {
				return &model.Department{ID: 1, Name: "Old Name", ParentID: nil}, nil
			},
			expectedError: nil,
		},
		{
			name:        "Ошибка: Департамент не найден",
			deptID:      404,
			newName:     "Non Existent",
			newParentID: nil,
			mockGetByID: func(ctx context.Context, id uint) (*model.Department, error) {
				return nil, errors.New("record not found")
			},
			expectedError: errors.New("record not found"),
		},
		{
			name:        "Ошибка: Цикл (департамент пытается стать родителем самого себя)",
			deptID:      5,
			newName:     "Self Parent",
			newParentID: uintPtr(5),
			mockGetByID: func(ctx context.Context, id uint) (*model.Department, error) {
				return &model.Department{ID: 5, Name: "Self Parent"}, nil
			},
			expectedError: errors.New("department cannot be its own parent"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockDepartmentRepository{
				getByIDFunc: tt.mockGetByID,
				updateFunc: func(ctx context.Context, dept *model.Department) error {
					return nil
				},
			}

			svc := NewDepartmentService(mockRepo)

			_, err := svc.Update(ctx, tt.deptID, tt.newName, tt.newParentID)

			if (err == nil && tt.expectedError != nil) || (err != nil && tt.expectedError == nil) {
				t.Fatalf("ожидали ошибку %v, но получили %v", tt.expectedError, err)
			}

			if err != nil && tt.expectedError != nil && err.Error() != tt.expectedError.Error() {
				t.Fatalf("ожидали ошибку '%v', но получили '%v'", tt.expectedError.Error(), err.Error())
			}
		})
	}
}

// хелпер что бы получить & на uint
func uintPtr(n uint) *uint {
	return &n
}
