package service

import "errors"

var (
	ErrDepartmentNotFound = errors.New("department not found")
	ErrCycleDetected      = errors.New("cyclical dependency detected: cannot set sub-department as parent")
	ErrParentNotFound     = errors.New("parent department not found")

	ErrEmployeeNotFound       = errors.New("employee not found")
	ErrInvalidDeleteMode      = errors.New("mode must be 'cascade' or 'reassign'")
	ErrReassignTargetNotFound = errors.New("reassign target department not found")
)
