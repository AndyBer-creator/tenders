package handlers_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tenders/db"
	"tenders/internal/handlers"

	"github.com/stretchr/testify/require"
)

// Интерфейс с методами, которые нам нужны для мокирования
type StorageMock interface {
	GetEmployeeByUsername(ctx context.Context, username string) (*db.Employee, error)
	IsUserResponsibleForOrganization(ctx context.Context, userID, organizationID int) (bool, error)
	CreateTender(ctx context.Context, tender *db.Tender) error
}

// Mock структура, реализующая StorageMock интерфейс
type MockStorage struct {
	employee        *db.Employee
	responsible     bool
	createTenderErr error
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

// Адаптер для соответствия *db.Storage
type MockStorageAdapter struct {
	mock StorageMock
}

// Реализация методов *db.Storage через адаптер и мок

func (a *MockStorageAdapter) GetEmployeeByUsername(ctx context.Context, username string) (*db.Employee, error) {
	return a.mock.GetEmployeeByUsername(ctx, username)
}

func (a *MockStorageAdapter) IsUserResponsibleForOrganization(ctx context.Context, userID, organizationID int) (bool, error) {
	return a.mock.IsUserResponsibleForOrganization(ctx, userID, organizationID)
}

func (a *MockStorageAdapter) CreateTender(ctx context.Context, tender *db.Tender) error {
	return a.mock.CreateTender(ctx, tender)
}

// Другие методы *db.Storage могут быть реализованы, если понадобятся в тестах

func TestPingHandler(t *testing.T) {
	handler := handlers.NewHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	w := httptest.NewRecorder()

	handler.PingHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Equal(t, "ok", string(body))
}

func TestCreateTenderHandler(t *testing.T) {
	testCases := []struct {
		name            string
		mockStorage     *MockStorage
		requestBody     string
		expectedCode    int
		expectedContent string
	}{
		{
			name: "success",
			mockStorage: &MockStorage{
				employee:    &db.Employee{ID: 1},
				responsible: true,
			},
			requestBody: `{
				"name": "Test Tender",
				"description": "Description",
				"serviceType": "Construction",
				"organizationId": 1,
				"creatorUsername": "user1"
			}`,
			expectedCode:    http.StatusOK,
			expectedContent: "Test Tender",
		},
		{
			name:         "invalid json",
			mockStorage:  &MockStorage{},
			requestBody:  `invalid json`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "missing fields",
			mockStorage:  &MockStorage{},
			requestBody:  `{"name":""}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "user not found",
			mockStorage: &MockStorage{
				employee: nil,
			},
			requestBody: `{
                "name": "Test Tender",
                "description": "Description",
                "serviceType": "Construction",
                "organizationId": 1,
                "creatorUsername": "user1"
            }`,
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "user not responsible",
			mockStorage: &MockStorage{
				employee:    &db.Employee{ID: 1},
				responsible: false,
			},
			requestBody: `{
                "name": "Test Tender",
                "description": "Description",
                "serviceType": "Construction",
                "organizationId": 1,
                "creatorUsername": "user1"
            }`,
			expectedCode: http.StatusForbidden,
		},
		{
			name: "create error",
			mockStorage: &MockStorage{
				employee:        &db.Employee{ID: 1},
				responsible:     true,
				createTenderErr: errors.New("db error"),
			},
			requestBody: `{
                "name": "Test Tender",
                "description": "Description",
                "serviceType": "Construction",
                "organizationId": 1,
                "creatorUsername": "user1"
            }`,
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &MockStorageAdapter{mock: tt.mockStorage}
			handler := handlers.NewHandler(adapter)

			req := httptest.NewRequest(http.MethodPost, "/api/tenders/new", strings.NewReader(tt.requestBody))
			w := httptest.NewRecorder()

			handler.CreateTenderHandler(w, req)

			res := w.Result()
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			require.Equal(t, tt.expectedCode, res.StatusCode)
			if tt.expectedContent != "" {
				require.Contains(t, string(body), tt.expectedContent)
			}
		})
	}
}
