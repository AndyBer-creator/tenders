package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"tenders/db"
)

// Handler оборачивает Storage для доступа к данным
type Handler struct {
	Store *db.Storage
}

// NewHandler создает новый Handler
func NewHandler(store *db.Storage) *Handler {
	return &Handler{Store: store}
}

// PingHandler отвечает "ok" для проверки сервера
func (h *Handler) PingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// CreateTenderHandler обрабатывает POST /api/tenders/new запрос
func (h *Handler) CreateTenderHandler(w http.ResponseWriter, r *http.Request) {
	// Ограничение размера тела, чтобы избежать DoS
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var tender db.Tender
	if err := json.Unmarshal(body, &tender); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Валидация полей согласно OpenAPI (можно расширить)
	if err := validateTenderRequest(&tender); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Статус должен быть "Created" при создании (по требованиям)
	tender.Status = "Created"
	// Версия устанавливается в CreateTender (1), так что не надо менять

	if err := h.Store.CreateTender(r.Context(), &tender); err != nil {
		http.Error(w, "Failed to create tender", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tender)
}

// validateTenderRequest проверяет необходимые поля по спецификации
func validateTenderRequest(t *db.Tender) error {
	if t.Name == "" || len(t.Name) > 100 {
		return errors.New("name is required and max length 100")
	}
	if t.Description == "" || len(t.Description) > 500 {
		return errors.New("description is required and max length 500")
	}
	switch t.ServiceType {
	case "Construction", "Delivery", "Manufacture":
		// ok
	default:
		return errors.New("invalid serviceType")
	}
	if t.OrganizationID <= 0 {
		return errors.New("organizationId must be positive")
	}
	if t.Status != "" && t.Status != "Created" {
		return errors.New("status must be 'Created' on creation")
	}
	return nil
}
