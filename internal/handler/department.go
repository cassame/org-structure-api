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

	if strings.TrimSpace(input.Name) == "" {
		respondWithError(w, http.StatusBadRequest, "Name is required")
		return
	}

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

	if strings.TrimSpace(input.Name) == "" {
		respondWithError(w, http.StatusBadRequest, "Name cannot be empty")
		return
	}

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

	err = h.service.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrDepartmentNotFound) {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Department deleted successfully"})
}
