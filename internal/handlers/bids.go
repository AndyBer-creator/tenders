package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"tenders/db"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) CreateBidHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var bid db.Bid
	if err := json.Unmarshal(body, &bid); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if err := validateBidRequest(&bid); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	bid.Status = "Created" // Статус при создании

	if err := h.Store.CreateBid(r.Context(), &bid); err != nil {
		http.Error(w, "Failed to create bid", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bid)
}

func validateBidRequest(b *db.Bid) error {
	if b.Name == "" || len(b.Name) > 100 {
		return errors.New("name is required and max length 100")
	}
	if b.Description == "" || len(b.Description) > 500 {
		return errors.New("description is required and max length 500")
	}
	if b.TenderID <= 0 {
		return errors.New("tenderId must be positive")
	}

	if b.OrganizationID <= 0 {
		return errors.New("organizationId must be positive")
	}
	if b.CreatorUsername == "" {
		return errors.New("creatorUsername is required")
	}
	if b.Status != "" && b.Status != "Created" {
		return errors.New("status must be 'Created' on creation")
	}
	return nil
}

func (h *Handler) GetUserBidsHandler(w http.ResponseWriter, r *http.Request) {
	params := parsePaginationParams(r)

	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Missing username parameter", http.StatusBadRequest)
		return
	}
	username = strings.TrimSpace(username)

	bids, err := h.Store.GetUserBids(r.Context(), username, params.Limit, params.Offset)
	if err != nil {
		http.Error(w, "Failed to get user bids", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bids)
}

func (h *Handler) GetBidsForTenderHandler(w http.ResponseWriter, r *http.Request) {
	params := parsePaginationParams(r)

	tenderIDStr := chi.URLParam(r, "tenderId")
	tenderID, err := strconv.Atoi(tenderIDStr)
	if err != nil {
		http.Error(w, "Invalid tenderId", http.StatusBadRequest)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Missing username parameter", http.StatusBadRequest)
		return
	}

	bids, err := h.Store.GetBidsForTender(r.Context(), tenderID, username, params.Limit, params.Offset)
	if err != nil {
		http.Error(w, "Failed to get bids for tender", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bids)
}

func (h *Handler) EditBidHandler(w http.ResponseWriter, r *http.Request) {
	bidIDStr := chi.URLParam(r, "bidId")
	bidID, err := strconv.Atoi(bidIDStr)
	if err != nil || bidID <= 0 {
		http.Error(w, "Invalid bidId", http.StatusBadRequest)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Missing username parameter", http.StatusBadRequest)
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
		Status      *string `json:"status"` // если нужно менять статус тут
	}

	if err := json.Unmarshal(body, &input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Получаем предложение из БД
	bid, err := h.Store.GetBid(r.Context(), bidID)
	if err != nil {
		http.Error(w, "Bid not found", http.StatusNotFound)
		return
	}

	// Получаем пользователя по username
	employee, err := h.Store.GetEmployeeByUsername(r.Context(), username)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Проверяем, что пользователь автор или ответственный за организацию предложения
	if employee.Username != bid.CreatorUsername {
		isResponsible, err := h.Store.IsUserResponsibleForOrganization(r.Context(), employee.ID, bid.OrganizationID)
		if err != nil || !isResponsible {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	}

	// Обновляем поля, если они переданы
	if input.Name != nil {
		if len(*input.Name) == 0 || len(*input.Name) > 100 {
			http.Error(w, "Invalid name length", http.StatusBadRequest)
			return
		}
		bid.Name = *input.Name
	}
	if input.Description != nil {
		if len(*input.Description) == 0 || len(*input.Description) > 500 {
			http.Error(w, "Invalid description length", http.StatusBadRequest)
			return
		}
		bid.Description = *input.Description
	}
	if input.Status != nil {
		// Если хотите разрешить изменение статуса через этот хендлер, добавьте проверку статуса
		// bid.Status = *input.Status
	}

	// Обновляем версию и дату обновления в методе UpdateBid
	if err := h.Store.UpdateBid(r.Context(), bid); err != nil {
		http.Error(w, "Failed to update bid", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bid)
}

func (h *Handler) UpdateBidStatusHandler(w http.ResponseWriter, r *http.Request) {
	bidIDStr := chi.URLParam(r, "bidId")
	bidID, err := strconv.Atoi(bidIDStr)
	if err != nil || bidID <= 0 {
		http.Error(w, "Invalid bidId", http.StatusBadRequest)
		return
	}

	status := r.URL.Query().Get("status")
	username := r.URL.Query().Get("username")
	if status == "" || username == "" {
		http.Error(w, "Missing status or username", http.StatusBadRequest)
		return
	}

	validStatuses := map[string]bool{"Created": true, "Published": true, "Canceled": true, "Approved": true, "Rejected": true}
	if !validStatuses[status] {
		http.Error(w, "Invalid status value", http.StatusBadRequest)
		return
	}

	// Проверяем пользователя и права
	employee, err := h.Store.GetEmployeeByUsername(r.Context(), username)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	bid, err := h.Store.GetBid(r.Context(), bidID)
	if err != nil {
		http.Error(w, "Bid not found", http.StatusNotFound)
		return
	}

	isResponsible, err := h.Store.IsUserResponsibleForOrganization(r.Context(), employee.ID, bid.OrganizationID)
	if err != nil || !isResponsible {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	bid.Status = status
	if err := h.Store.UpdateBid(r.Context(), bid); err != nil {
		http.Error(w, "Failed to update bid status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bid)
}
func (h *Handler) RollbackBidHandler(w http.ResponseWriter, r *http.Request) {
	bidIDStr := chi.URLParam(r, "bidId")
	versionStr := chi.URLParam(r, "version")

	bidID, err1 := strconv.Atoi(bidIDStr)
	version, err2 := strconv.Atoi(versionStr)
	if err1 != nil || err2 != nil || bidID <= 0 || version < 1 {
		http.Error(w, "Invalid bidId or version", http.StatusBadRequest)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Missing username", http.StatusBadRequest)
		return
	}

	employee, err := h.Store.GetEmployeeByUsername(r.Context(), username)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	currentBid, err := h.Store.GetBid(r.Context(), bidID)
	if err != nil {
		http.Error(w, "Bid not found", http.StatusNotFound)
		return
	}

	isResponsible, err := h.Store.IsUserResponsibleForOrganization(r.Context(), employee.ID, currentBid.OrganizationID)
	if err != nil || !isResponsible {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	versionBid, err := h.Store.GetBidVersion(r.Context(), bidID, version)
	if err != nil {
		http.Error(w, "Version not found", http.StatusNotFound)
		return
	}

	// Откатываем значения
	currentBid.Name = versionBid.Name
	currentBid.Description = versionBid.Description
	currentBid.Status = versionBid.Status
	currentBid.Version++ // Инкрементируем версию

	if err := h.Store.UpdateBid(r.Context(), currentBid); err != nil {
		http.Error(w, "Failed to rollback bid", http.StatusInternalServerError)
		return
	}

	// Сохраняем новую версию после отката
	_ = h.Store.SaveBidVersion(r.Context(), currentBid)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currentBid)
}

func (h *Handler) SubmitBidDecisionHandler(w http.ResponseWriter, r *http.Request) {
	bidIDStr := chi.URLParam(r, "bidId")
	decision := r.URL.Query().Get("decision")
	username := r.URL.Query().Get("username")

	if bidIDStr == "" || decision == "" || username == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	bidID, err := strconv.Atoi(bidIDStr)
	if err != nil || (decision != "Approved" && decision != "Rejected") {
		http.Error(w, "Invalid bidId or decision", http.StatusBadRequest)
		return
	}

	employee, err := h.Store.GetEmployeeByUsername(r.Context(), username)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	bid, err := h.Store.GetBid(r.Context(), bidID)
	if err != nil {
		http.Error(w, "Bid not found", http.StatusNotFound)
		return
	}

	isResponsible, err := h.Store.IsUserResponsibleForOrganization(r.Context(), employee.ID, bid.OrganizationID)
	if err != nil || !isResponsible {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Добавляем решение текущего пользователя в хранилище
	err = h.Store.AddBidDecision(r.Context(), bid.ID, employee.ID, decision)
	if err != nil {
		http.Error(w, "Failed to submit decision", http.StatusInternalServerError)
		return
	}

	// Подсчитываем количество одобрений и отклонений
	accepts, rejects, err := h.Store.GetBidDecisionsCount(r.Context(), bid.ID)
	if err != nil {
		http.Error(w, "Failed to get decision counts", http.StatusInternalServerError)
		return
	}

	// Получаем количество ответственных за организацию
	respCount, err := h.Store.GetResponsibleCount(r.Context(), bid.OrganizationID)
	if err != nil {
		http.Error(w, "Failed to get responsible count", http.StatusInternalServerError)
		return
	}

	quorum := respCount
	if quorum > 3 {
		quorum = 3
	}

	// Логика постановки статуса согласно кворуму
	switch {
	case rejects > 0:
		bid.Status = "Rejected"
	case accepts >= quorum:
		bid.Status = "Approved"

		tender, err := h.Store.GetTender(r.Context(), bid.TenderID)
		if err == nil {
			tender.Status = "Closed"
			_ = h.Store.UpdateTender(r.Context(), tender)
		}
	default:
		// Статус остаётся прежним, например, "Published"
	}

	err = h.Store.UpdateBid(r.Context(), bid)
	if err != nil {
		http.Error(w, "Failed to update bid status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bid)
}

func (h *Handler) GetBidReviewsHandler(w http.ResponseWriter, r *http.Request) {
	tenderIDStr := chi.URLParam(r, "tenderId")
	authorUsername := r.URL.Query().Get("authorUsername")
	requesterUsername := r.URL.Query().Get("requesterUsername")

	if tenderIDStr == "" || authorUsername == "" || requesterUsername == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	tenderID, err := strconv.Atoi(tenderIDStr)
	if err != nil {
		http.Error(w, "Invalid tenderId", http.StatusBadRequest)
		return
	}

	requester, err := h.Store.GetEmployeeByUsername(r.Context(), requesterUsername)
	if err != nil {
		http.Error(w, "Requester user not found", http.StatusUnauthorized)
		return
	}

	// Проверка прав: запросивший должен быть ответственным за организацию тендера
	tender, err := h.Store.GetTender(r.Context(), tenderID)
	if err != nil {
		http.Error(w, "Tender not found", http.StatusNotFound)
		return
	}

	isResponsible, err := h.Store.IsUserResponsibleForOrganization(r.Context(), requester.ID, tender.OrganizationID)
	if err != nil || !isResponsible {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Получаем отзывы по предложениям автора, для указанного тендера
	reviews, err := h.Store.GetBidReviewsByAuthorForTender(r.Context(), authorUsername, tenderID)
	if err != nil {
		http.Error(w, "Failed to get reviews", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reviews)
}
func (h *Handler) CreateBidFeedbackHandler(w http.ResponseWriter, r *http.Request) {
	bidIDStr := chi.URLParam(r, "bidId")
	username := r.URL.Query().Get("username")
	feedback := r.URL.Query().Get("bidFeedback")

	bidID, err := strconv.Atoi(bidIDStr)
	if err != nil {
		http.Error(w, "Invalid bidId", http.StatusBadRequest)
		return
	}

	if username == "" || feedback == "" {
		http.Error(w, "Missing username or feedback", http.StatusBadRequest)
		return
	}

	employee, err := h.Store.GetEmployeeByUsername(r.Context(), username)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	bid, err := h.Store.GetBid(r.Context(), bidID)
	if err != nil {
		http.Error(w, "Bid not found", http.StatusNotFound)
		return
	}

	isResponsible, err := h.Store.IsUserResponsibleForOrganization(r.Context(), employee.ID, bid.OrganizationID)
	if err != nil || !isResponsible {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	review := &db.BidReview{
		BidID:       bidID,
		Description: feedback,
	}

	err = h.Store.CreateBidReview(r.Context(), review)
	if err != nil {
		http.Error(w, "Failed to submit feedback", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(review)
}

// Другие обработчики: редактирование, статус, откат версии, решения по предложению и отзывы можно сделать по аналогии с тендерами
