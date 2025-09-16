package handlers_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"tenders/db"
	"tenders/internal/handlers"
	"tenders/internal/handlers/testutils"
	"testing"

	"github.com/stretchr/testify/require"
)

// MockStorage реализует StorageInterface
type MockStorage struct {
	employee             *db.Employee
	responsible          bool
	createTenderErr      error
	GetTendersFunc       func(ctx context.Context, serviceTypes []string, limit, offset int) ([]db.Tender, error)
	GetBidFunc           func(ctx context.Context, bidID int) (*db.Bid, error)
	GetUserBidsFunc      func(ctx context.Context, username string, limit, offset int) ([]db.Bid, error)
	GetBidsForTenderFunc func(ctx context.Context, tenderID int, username string, limit, offset int) ([]db.Bid, error)
}

func (m *MockStorage) GetEmployeeByUsername(ctx context.Context, username string) (*db.Employee, error) {
	if m.employee == nil {
		return nil, errors.New("not found")
	}
	return m.employee, nil
}
func (m *MockStorage) IsUserResponsibleForOrganization(ctx context.Context, userID, organizationID int) (bool, error) {
	return m.responsible, nil
}

func (m *MockStorage) CreateTender(ctx context.Context, tender *db.Tender) error {
	return m.createTenderErr
}

func (m *MockStorage) GetTender(ctx context.Context, tenderID int) (*db.Tender, error) {
	return &db.Tender{ID: tenderID, Name: "Test Tender", OrganizationID: 1}, nil
}

func (m *MockStorage) UpdateTender(ctx context.Context, tender *db.Tender) error      { return nil }
func (m *MockStorage) SaveTenderVersion(ctx context.Context, tender *db.Tender) error { return nil }
func (m *MockStorage) GetTenderVersion(ctx context.Context, tenderID int, version int) (*db.Tender, error) {
	return &db.Tender{ID: tenderID, Name: "Tender Version"}, nil
}
func (m *MockStorage) GetTenders(ctx context.Context, serviceTypes []string, limit, offset int) ([]db.Tender, error) {
	if m.GetTendersFunc != nil {
		return m.GetTendersFunc(ctx, serviceTypes, limit, offset)
	}
	return []db.Tender{{ID: 1, Name: "Sample Tender"}}, nil
}

func (m *MockStorage) GetUserTenders(ctx context.Context, username string, limit, offset int) ([]db.Tender, error) {
	return []db.Tender{
		{ID: 1, Name: "User Tender"},
	}, nil
}

func (m *MockStorage) CreateBid(ctx context.Context, bid *db.Bid) error { return nil }
func (m *MockStorage) GetBid(ctx context.Context, bidID int) (*db.Bid, error) {
	if m.GetBidFunc != nil {
		return m.GetBidFunc(ctx, bidID)
	}
	return &db.Bid{
		ID:              bidID,
		Name:            "Test Bid",
		Description:     "Bid Description",
		Status:          "Created",
		TenderID:        1,
		OrganizationID:  1,
		CreatorUsername: "user1",
		Version:         1,
	}, nil
}
func (m *MockStorage) UpdateBid(ctx context.Context, bid *db.Bid) error { return nil }
func (m *MockStorage) GetUserBids(ctx context.Context, username string, limit, offset int) ([]db.Bid, error) {
	if m.GetUserBidsFunc != nil {
		return m.GetUserBidsFunc(ctx, username, limit, offset)
	}
	return []db.Bid{
		{
			ID:              1,
			Name:            "User Bid",
			Description:     "Description for user bid",
			Status:          "Created",
			TenderID:        1,
			CreatorUsername: username,
			Version:         1,
		},
	}, nil
}
func (m *MockStorage) GetBidsForTender(ctx context.Context, tenderID int, username string, limit, offset int) ([]db.Bid, error) {
	if m.GetBidsForTenderFunc != nil {
		return m.GetBidsForTenderFunc(ctx, tenderID, username, limit, offset)
	}
	return []db.Bid{
		{
			ID:          2,
			Name:        "Tender Bid",
			Description: "Description for tender bid",
			Status:      "Published",
			TenderID:    tenderID,
			Version:     1,
		},
	}, nil
}
func (m *MockStorage) GetBidVersion(ctx context.Context, bidID, version int) (*db.Bid, error) {
	return &db.Bid{
		ID:              bidID,
		Name:            "Bid Version Name",
		Description:     "Bid Version Description",
		Status:          "Created",
		TenderID:        1,
		OrganizationID:  1,
		CreatorUsername: "user1",
		Version:         version,
	}, nil
}
func (m *MockStorage) SaveBidVersion(ctx context.Context, bid *db.Bid) error { return nil }

func (m *MockStorage) AddBidDecision(ctx context.Context, bidID, employeeID int, decision string) error {
	return nil
}
func (m *MockStorage) GetBidDecisionsCount(ctx context.Context, bidID int) (int, int, error) {
	return 1, 0, nil
}
func (m *MockStorage) GetResponsibleCount(ctx context.Context, organizationID int) (int, error) {
	return 1, nil
}

func (m *MockStorage) GetBidReviewsByAuthorForTender(ctx context.Context, authorUsername string, tenderID int) ([]db.BidReview, error) {
	return []db.BidReview{{ID: 1, Description: "Good"}}, nil
}
func (m *MockStorage) CreateBidReview(ctx context.Context, review *db.BidReview) error { return nil }

func TestGetTendersHandler(t *testing.T) {
	mockStore := &MockStorage{}
	handler := handlers.NewHandler(mockStore)

	req := httptest.NewRequest("GET", "/api/tenders", nil)
	w := httptest.NewRecorder()

	handler.GetTendersHandler(w, req)

	res := w.Result()
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	require.Equal(t, 200, res.StatusCode)
	require.Contains(t, string(body), "Sample Tender")
}

func TestCreateTenderHandler(t *testing.T) {
	mockStore := &MockStorage{
		employee:    &db.Employee{ID: 1},
		responsible: true,
	}
	handler := handlers.NewHandler(mockStore)

	reqBody := `{
        "name": "Test Tender",
        "description": "Desc",
        "serviceType": "Construction",
        "organizationId": 1,
        "creatorUsername": "user1"
    }`
	req := httptest.NewRequest(http.MethodPost, "/api/tenders/new?username=user1", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateTenderHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Contains(t, string(body), "Test Tender")
}

func TestChangeTenderStatusHandler(t *testing.T) {
	mockStore := &MockStorage{}
	handler := handlers.NewHandler(mockStore)

	req := httptest.NewRequest(http.MethodPut, "/api/tenders/123/status?username=user1", strings.NewReader(`{"status":"closed"}`))
	req.Header.Set("Content-Type", "application/json")
	req = testutils.WithChiURLParams(req, map[string]string{"tenderId": "123"})

	w := httptest.NewRecorder()

	handler.ChangeTenderStatusHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Contains(t, string(body), "status")
}

func TestUpdateTenderStatusHandler(t *testing.T) {
	mockStore := &MockStorage{}
	handler := handlers.NewHandler(mockStore)

	reqBody := `{"status":"closed"}`
	req := httptest.NewRequest(http.MethodPut, "/api/tenders/123/status?username=user1", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req = testutils.WithChiURLParams(req, map[string]string{"tenderId": "123"})

	w := httptest.NewRecorder()
	handler.ChangeTenderStatusHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Contains(t, string(body), "status")
}

func TestRollbackTenderHandler(t *testing.T) {
	mockStore := &MockStorage{}
	handler := handlers.NewHandler(mockStore)

	req := httptest.NewRequest(http.MethodPut, "/api/tenders/123/rollback/1?username=user1", nil)
	req = testutils.WithChiURLParams(req, map[string]string{"tenderId": "123", "version": "1"})

	w := httptest.NewRecorder()
	handler.RollbackTenderHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Contains(t, string(body), "rollback")
}

func TestEditTenderHandler(t *testing.T) {
	mockStore := &MockStorage{
		employee:    &db.Employee{ID: 1},
		responsible: true,
	}
	handler := handlers.NewHandler(mockStore)

	reqBody := `{"name":"Updated Tender"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/tenders/123/edit?username=user1", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req = testutils.WithChiURLParams(req, map[string]string{"tenderId": "123"})

	w := httptest.NewRecorder()
	handler.EditTenderHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Contains(t, string(body), "Updated Tender")
}

func TestCreateBidHandler(t *testing.T) {
	mockStore := &MockStorage{}
	handler := handlers.NewHandler(mockStore)

	reqBody := `{
        "tenderId": 1,
        "bidderUsername": "user1",
        "name": "Bid Name"
    }`
	req := httptest.NewRequest(http.MethodPost, "/api/bids/new?username=user1", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateBidHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Contains(t, string(body), "bid")
}

func TestGetUserBidsHandler(t *testing.T) {
	mockStore := &MockStorage{}
	handler := handlers.NewHandler(mockStore)

	req := httptest.NewRequest(http.MethodGet, "/api/bids/my?username=user1", nil)
	w := httptest.NewRecorder()

	handler.GetUserBidsHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Contains(t, string(body), "bids")
}

func TestGetBidsForTenderHandler(t *testing.T) {
	mockStore := &MockStorage{}
	handler := handlers.NewHandler(mockStore)

	req := httptest.NewRequest(http.MethodPut, "/api/bids/1/list?username=user1", nil)
	req = testutils.WithChiURLParams(req, map[string]string{"tenderId": "1"})

	w := httptest.NewRecorder()

	handler.GetBidsForTenderHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Contains(t, string(body), "bids")
}

func TestEditBidHandler(t *testing.T) {
	mockStore := &MockStorage{}
	handler := handlers.NewHandler(mockStore)

	reqBody := `{"name": "Updated Bid"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/bids/1/edit?username=user1", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req = testutils.WithChiURLParams(req, map[string]string{"bidId": "1"})

	w := httptest.NewRecorder()

	handler.EditBidHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Contains(t, string(body), "Updated Bid")
}

func TestUpdateBidStatusHandler(t *testing.T) {
	mockStore := &MockStorage{}
	handler := handlers.NewHandler(mockStore)

	reqBody := `{"status":"accepted"}`
	req := httptest.NewRequest(http.MethodPut, "/api/bids/1/status?username=user1", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req = testutils.WithChiURLParams(req, map[string]string{"bidId": "1"})

	w := httptest.NewRecorder()

	handler.UpdateBidStatusHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Contains(t, string(body), "status")
}

func TestRollbackBidHandler(t *testing.T) {
	mockStore := &MockStorage{}
	handler := handlers.NewHandler(mockStore)

	req := httptest.NewRequest(http.MethodPut, "/api/bids/1/rollback/1?username=user1", nil)
	req = testutils.WithChiURLParams(req, map[string]string{"bidId": "1", "version": "1"})

	w := httptest.NewRecorder()

	handler.RollbackBidHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Contains(t, string(body), "rollback")
}

func TestSubmitBidDecisionHandler(t *testing.T) {
	mockStore := &MockStorage{}
	handler := handlers.NewHandler(mockStore)

	reqBody := `{"decision":"accept"}`
	req := httptest.NewRequest(http.MethodPut, "/api/bids/1/submit_decision?username=user1", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req = testutils.WithChiURLParams(req, map[string]string{"bidId": "1"})

	w := httptest.NewRecorder()

	handler.SubmitBidDecisionHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Contains(t, string(body), "decision")
}

func TestGetBidReviewsHandler(t *testing.T) {
	mockStore := &MockStorage{}
	handler := handlers.NewHandler(mockStore)

	req := httptest.NewRequest(http.MethodGet, "/api/bids/1/reviews?username=user1", nil)
	req = testutils.WithChiURLParams(req, map[string]string{"tenderId": "1"})

	w := httptest.NewRecorder()

	handler.GetBidReviewsHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Contains(t, string(body), "reviews")
}

func TestCreateBidFeedbackHandler(t *testing.T) {
	mockStore := &MockStorage{
		employee: &db.Employee{ID: 1}, // Эмуляция успешного поиска пользователя
	}
	handler := handlers.NewHandler(mockStore)

	req := httptest.NewRequest(http.MethodPut, "/api/bids/1/feedback?username=user1&bidFeedback=good", nil)
	req = testutils.WithChiURLParams(req, map[string]string{"bidId": "1"})

	w := httptest.NewRecorder()

	handler.CreateBidFeedbackHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Contains(t, string(body), "feedback")
}
