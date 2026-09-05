package httpadapter

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
	"github.com/go-chi/chi/v5"
)

type TemplateGroupHandler struct {
	service ports.TemplateGroupService
}

func NewTemplateGroupHandler(service *ports.TemplateGroupService) *TemplateGroupHandler {
	return &TemplateGroupHandler{service: *service}
}

func (h *TemplateGroupHandler) GetTemplateGroups(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized: Invalid user ID session", http.StatusUnauthorized)
		return
	}

	groups, err := h.service.GetTemplateGroups(userID)
	if err != nil {
		log.Printf("Error fetching template groups: %v", err)
		http.Error(w, "Could not fetch template groups", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(groups)
}

func (h *TemplateGroupHandler) CreateTemplateGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized: Invalid user ID session", http.StatusUnauthorized)
		return
	}

	var group domain.TemplateGroup
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		log.Printf("JSON decode error: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	group.UserID = userID

	created, err := h.service.CreateTemplateGroup(userID, group)
	if err != nil {
		log.Printf("Error creating template group: %v", err)
		switch err {
		case domain.ErrMissingTemplateGroup:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "Error creating template group", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *TemplateGroupHandler) UpdateTemplateGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized: Invalid user ID session", http.StatusUnauthorized)
		return
	}

	groupIDStr := chi.URLParam(r, "id")
	if groupIDStr == "" {
		http.Error(w, "Missing template group ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(groupIDStr)
	if err != nil {
		http.Error(w, "Id is not valid", http.StatusBadRequest)
		return
	}

	var group domain.TemplateGroup
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		log.Printf("JSON decode error: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	group.ID = id
	group.UserID = userID

	if err := h.service.UpdateTemplateGroup(userID, group); err != nil {
		log.Printf("Error updating template group %d for user %d: %v", id, userID, err)
		switch err {
		case domain.ErrTemplateGroupNotFound, domain.ErrMissingTemplateGroup:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case domain.ErrUnauthorized:
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		default:
			http.Error(w, "Could not update template group", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TemplateGroupHandler) DeleteTemplateGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized: Invalid user ID session", http.StatusUnauthorized)
		return
	}

	groupIDStr := chi.URLParam(r, "id")
	if groupIDStr == "" {
		http.Error(w, "Missing template group ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(groupIDStr)
	if err != nil {
		http.Error(w, "Id is not valid", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteTemplateGroup(userID, id); err != nil {
		log.Printf("Error deleting template group: %v", err)
		switch err {
		case domain.ErrTemplateGroupNotFound:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case domain.ErrUnauthorized:
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		default:
			http.Error(w, "Error deleting template group", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
