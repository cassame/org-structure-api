package handler

import (
	"apiOrganizationStructure/internal/service"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type EmployeeHandler struct {
	service *service.EmployeeService
}

func NewEmployeeHandler(s *service.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{service: s}
}

type employeeInput struct {
	DepartmentID uint    `json:"department_id"`
	FullName     string  `json:"full_name"`
	Position     string  `json:"position"`
	HiredAt      *string `json:"hired_at"`
}

// POST /api/employees
func (h *EmployeeHandler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	var input employeeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if strings.TrimSpace(input.FullName) == "" || strings.TrimSpace(input.Position) == "" || input.DepartmentID == 0 {
		respondWithError(w, http.StatusBadRequest, "full_name, position, and department_id are required")
		return
	}

	var hiredAtTime *time.Time
	if input.HiredAt != nil && *input.HiredAt != "" {
		parsedTime, err := time.Parse("2006-01-02", *input.HiredAt)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid date format for hired_at. Use YYYY-MM-DD")
			return
		}
		hiredAtTime = &parsedTime
	}

	emp, err := h.service.Create(r.Context(), input.DepartmentID, input.FullName, input.Position, hiredAtTime)
	if err != nil {
		if errors.Is(err, service.ErrParentNotFound) {
			respondWithError(w, http.StatusBadRequest, "Department not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Internal server error: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, emp)
}

func (h *EmployeeHandler) HandleEmployeeByID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		respondWithError(w, http.StatusBadRequest, "Missing employee ID")
		return
	}

	idParam := parts[2]
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid employee ID")
		return
	}

	switch r.Method {
	case http.MethodPatch:
		h.UpdateEmployee(w, r, uint(id))
	case http.MethodDelete:
		h.DeleteEmployee(w, r, uint(id))
	default:
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *EmployeeHandler) UpdateEmployee(w http.ResponseWriter, r *http.Request, id uint) {
	var input employeeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if strings.TrimSpace(input.FullName) == "" || strings.TrimSpace(input.Position) == "" || input.DepartmentID == 0 {
		respondWithError(w, http.StatusBadRequest, "fields cannot be empty")
		return
	}

	var hiredAtTime *time.Time
	if input.HiredAt != nil && *input.HiredAt != "" {
		parsedTime, err := time.Parse("2006-01-02", *input.HiredAt)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
			return
		}
		hiredAtTime = &parsedTime
	}

	emp, err := h.service.Update(r.Context(), id, input.DepartmentID, input.FullName, input.Position, hiredAtTime)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmployeeNotFound):
			respondWithError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrParentNotFound):
			respondWithError(w, http.StatusBadRequest, "Target department not found")
		default:
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, emp)
}

func (h *EmployeeHandler) DeleteEmployee(w http.ResponseWriter, r *http.Request, id uint) {
	err := h.service.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrEmployeeNotFound) {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Employee deleted successfully"})
}
