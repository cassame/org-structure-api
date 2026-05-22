package handler

import (
	"apiOrganizationStructure/internal/service"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type DepartmentHandler struct {
	service *service.DepartmentService
}

func NewDepartmentHandler(s *service.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{service: s}
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func (h *DepartmentHandler) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		ParentID *uint  `json:"parent_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	name, err := validateName(input.Name)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	input.Name = name

	dept, err := h.service.Create(r.Context(), input.Name, input.ParentID)
	if err != nil {
		if errors.Is(err, service.ErrParentNotFound) {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Internal server error: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, dept)
}

func (h *DepartmentHandler) GetTree(w http.ResponseWriter, r *http.Request) {
	withEmployees := r.URL.Query().Get("with_employees") == "true"

	tree, err := h.service.GetTree(r.Context(), withEmployees)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	respondWithJSON(w, http.StatusOK, tree)
}

func (h *DepartmentHandler) UpdateDepartment(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid department ID")
		return
	}
	id := uint(id64)

	var input struct {
		Name     string `json:"name"`
		ParentID *uint  `json:"parent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	name, err := validateName(input.Name)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	input.Name = name

	dept, err := h.service.Update(r.Context(), id, input.Name, input.ParentID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrDepartmentNotFound):
			respondWithError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrCycleDetected) || errors.Is(err, service.ErrParentNotFound):
			respondWithError(w, http.StatusBadRequest, err.Error())
		default:
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, dept)
}

func (h *DepartmentHandler) DeleteDepartment(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid department ID")
		return
	}
	id := uint(id64)

	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = "cascade"
	}

	var reassignTo *uint
	if r.URL.Query().Get("reassign_to_department_id") != "" {
		val, err := strconv.ParseUint(r.URL.Query().Get("reassign_to_department_id"), 10, 32)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid reassign_to_department_id")
			return
		}
		u := uint(val)
		reassignTo = &u
	}

	err = h.service.Delete(r.Context(), id, mode, reassignTo)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrDepartmentNotFound):
			respondWithError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrReassignTargetNotFound):
			respondWithError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrInvalidDeleteMode):
			respondWithError(w, http.StatusBadRequest, err.Error())
		default:
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)

}

func (h *DepartmentHandler) GetDepartmentByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid department ID")
		return
	}

	depth := 1
	if d := r.URL.Query().Get("depth"); d != "" {
		parsed, err := strconv.Atoi(d)
		if err != nil || parsed < 1 || parsed > 5 {
			respondWithError(w, http.StatusBadRequest, "depth must be between 1 and 5")
			return
		}
		depth = parsed
	}

	withEmployees := r.URL.Query().Get("include_employees") != "false"

	dept, err := h.service.GetByID(r.Context(), uint(id64), depth, withEmployees)
	if err != nil {
		if errors.Is(err, service.ErrDepartmentNotFound) {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	respondWithJSON(w, http.StatusOK, dept)
}

func validateName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 200 {
		return "", errors.New("name must be between 1 and 200 characters")
	}
	return name, nil
}
