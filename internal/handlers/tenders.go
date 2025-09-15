package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type PaginationParams struct {
	Limit  int
	Offset int
}

// parsePaginationParams парсит limit и offset из query, с дефолтами и ограничениями
func parsePaginationParams(r *http.Request) PaginationParams {
	var params PaginationParams
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	params.Limit = 5 // дефолт
	params.Offset = 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			params.Limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			params.Offset = o
		}
	}
	return params
}

// GetTendersHandler возвращает список тендеров с фильтрами по типу serviceType
func (h *Handler) GetTendersHandler(w http.ResponseWriter, r *http.Request) {
	params := parsePaginationParams(r)

	// Фильтр service_type - может быть несколько через query param
	serviceTypes := r.URL.Query()["service_type"]
	// Проверим и отфильтруем дубликаты или неверные значения
	allowedTypes := map[string]bool{"Construction": true, "Delivery": true, "Manufacture": true}
	var filteredTypes []string
	for _, v := range serviceTypes {
		if allowedTypes[v] {
			filteredTypes = append(filteredTypes, v)
		}
	}

	tenders, err := h.Store.GetTenders(r.Context(), filteredTypes, params.Limit, params.Offset)
	if err != nil {
		http.Error(w, "Failed to get tenders", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tenders)
}

// GetUserTendersHandler возвращает список тендеров для пользователя username
func (h *Handler) GetUserTendersHandler(w http.ResponseWriter, r *http.Request) {
	params := parsePaginationParams(r)

	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Missing username parameter", http.StatusBadRequest)
		return
	}
	username = strings.TrimSpace(username)

	tenders, err := h.Store.GetUserTenders(r.Context(), username, params.Limit, params.Offset)
	if err != nil {
		http.Error(w, "Failed to get user tenders", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tenders)
}
func (h *Handler) UpdateTenderStatusHandler(w http.ResponseWriter, r *http.Request) {
	tenderIDStr := chi.URLParam(r, "tenderId")
	tenderID, err := strconv.Atoi(tenderIDStr)
	if err != nil || tenderID <= 0 {
		http.Error(w, "Invalid tenderId", http.StatusBadRequest)
		return
	}

	status := r.URL.Query().Get("status")
	username := r.URL.Query().Get("username")
	if status == "" || username == "" {
		http.Error(w, "Missing status or username", http.StatusBadRequest)
		return
	}

	// Проверить права пользователя (ответственный за организацию)
	employee, err := h.Store.GetEmployeeByUsername(r.Context(), username)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	tender, err := h.Store.GetTender(r.Context(), tenderID)
	if err != nil {
		http.Error(w, "Tender not found", http.StatusNotFound)
		return
	}

	// Проверка, что пользователь ответственность за организацию этого тендера
	isResponsible, err := h.Store.IsUserResponsibleForOrganization(r.Context(), employee.ID, tender.OrganizationID)
	if err != nil || !isResponsible {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Обновить статус тендера
	tender.Status = status
	err = h.Store.UpdateTender(r.Context(), tender)
	if err != nil {
		http.Error(w, "Failed to update status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tender)
}
func (h *Handler) EditTenderHandler(w http.ResponseWriter, r *http.Request) {
	tenderIDStr := chi.URLParam(r, "tenderId")
	tenderID, err := strconv.Atoi(tenderIDStr)
	if err != nil || tenderID <= 0 {
		http.Error(w, "Invalid tenderId", http.StatusBadRequest)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Missing username", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Cannot read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var input struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		ServiceType *string `json:"serviceType"`
	}

	if err := json.Unmarshal(body, &input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	employee, err := h.Store.GetEmployeeByUsername(r.Context(), username)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	tender, err := h.Store.GetTender(r.Context(), tenderID)
	if err != nil {
		http.Error(w, "Tender not found", http.StatusNotFound)
		return
	}

	isResponsible, err := h.Store.IsUserResponsibleForOrganization(r.Context(), employee.ID, tender.OrganizationID)
	if err != nil || !isResponsible {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if input.Name != nil {
		tender.Name = *input.Name
	}
	if input.Description != nil {
		tender.Description = *input.Description
	}
	if input.ServiceType != nil {
		tender.ServiceType = *input.ServiceType
	}

	if err := h.Store.UpdateTender(r.Context(), tender); err != nil {
		http.Error(w, "Failed to update tender", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tender)
}
func (h *Handler) RollbackTenderHandler(w http.ResponseWriter, r *http.Request) {
	tenderIDStr := chi.URLParam(r, "tenderId")
	versionStr := chi.URLParam(r, "version")

	tenderID, err := strconv.Atoi(tenderIDStr)
	if err != nil {
		http.Error(w, "invalid tender ID", http.StatusBadRequest)
		return
	}
	version, err := strconv.Atoi(versionStr)
	if err != nil || version < 1 {
		http.Error(w, "invalid version number", http.StatusBadRequest)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "missing username", http.StatusBadRequest)
		return
	}

	employee, err := h.Store.GetEmployeeByUsername(r.Context(), username)
	if err != nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	currentTender, err := h.Store.GetTender(r.Context(), tenderID)
	if err != nil {
		http.Error(w, "tender not found", http.StatusNotFound)
		return
	}

	isResponsible, err := h.Store.IsUserResponsibleForOrganization(r.Context(), employee.ID, currentTender.OrganizationID)
	if err != nil || !isResponsible {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	versionTender, err := h.Store.GetTenderVersion(r.Context(), tenderID, version)
	if err != nil {
		http.Error(w, "version not found", http.StatusNotFound)
		return
	}

	// Обновляем текущий тендер данными версии с инкрементом версии
	currentTender.Name = versionTender.Name
	currentTender.Description = versionTender.Description
	currentTender.ServiceType = versionTender.ServiceType
	currentTender.Status = versionTender.Status
	currentTender.Version++
	// Организация и CreatedAt менять не нужно

	err = h.Store.UpdateTender(r.Context(), currentTender)
	if err != nil {
		http.Error(w, "failed to update tender", http.StatusInternalServerError)
		return
	}

	// Сохраняем новую версию после отката
	err = h.Store.SaveTenderVersion(r.Context(), currentTender)
	if err != nil {
		// Логируем ошибку, но не прерываем
		// log.Printf("failed to save tender version: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currentTender)
}
func (h *Handler) ChangeTenderStatusHandler(w http.ResponseWriter, r *http.Request) {
	tenderIDStr := chi.URLParam(r, "tenderId")
	username := r.URL.Query().Get("username")
	newStatus := r.URL.Query().Get("status")

	if tenderIDStr == "" || username == "" || newStatus == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	tenderID, err := strconv.Atoi(tenderIDStr)
	if err != nil {
		http.Error(w, "Invalid tenderId", http.StatusBadRequest)
		return
	}

	// Проверка статуса
	allowedStatuses := map[string]bool{"CREATED": true, "PUBLISHED": true, "CLOSED": true}
	if !allowedStatuses[newStatus] {
		http.Error(w, "Invalid status value", http.StatusBadRequest)
		return
	}

	tender, err := h.Store.GetTender(r.Context(), tenderID)
	if err != nil {
		http.Error(w, "Tender not found", http.StatusNotFound)
		return
	}

	employee, err := h.Store.GetEmployeeByUsername(r.Context(), username)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Проверка прав ответственного
	isResponsible, err := h.Store.IsUserResponsibleForOrganization(r.Context(), employee.ID, tender.OrganizationID)
	if err != nil || !isResponsible {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Проверка возможности перехода статуса
	switch tender.Status {
	case "CREATED":
		if newStatus != "PUBLISHED" {
			http.Error(w, "Invalid status transition", http.StatusBadRequest)
			return
		}
	case "PUBLISHED":
		if newStatus != "CLOSED" {
			http.Error(w, "Invalid status transition", http.StatusBadRequest)
			return
		}
	case "CLOSED":
		http.Error(w, "Tender is already closed", http.StatusBadRequest)
		return
	}

	tender.Status = newStatus
	err = h.Store.UpdateTender(r.Context(), tender)
	if err != nil {
		http.Error(w, "Failed to update tender status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tender)
}
