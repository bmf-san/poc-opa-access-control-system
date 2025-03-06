package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bmf-san/poc-opa-access-control-system/internal/interfaces"
	"github.com/bmf-san/poc-opa-access-control-system/internal/model"
)

// mockDB implements interfaces.DBConn
type mockDB struct {
	QueryFunc    func(ctx context.Context, sql string, args ...interface{}) (interfaces.DBRows, error)
	QueryRowFunc func(ctx context.Context, sql string, args ...interface{}) interfaces.DBRow
}

type mockRows struct {
	data     [][]interface{}
	current  int
	scanFunc func(dest ...interface{}) error
}

func (m *mockRows) Close()     {}
func (m *mockRows) Err() error { return nil }
func (m *mockRows) Next() bool {
	if m.current >= len(m.data) {
		return false
	}
	m.current++
	return true
}
func (m *mockRows) Scan(dest ...interface{}) error {
	if m.current <= 0 || m.current > len(m.data) {
		return fmt.Errorf("invalid row access")
	}
	return m.scanFunc(dest...)
}

func (m *mockDB) Query(ctx context.Context, sql string, args ...interface{}) (interfaces.DBRows, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc(ctx, sql, args...)
	}
	return nil, fmt.Errorf("Query not implemented")
}

func (m *mockDB) QueryRow(ctx context.Context, sql string, args ...interface{}) interfaces.DBRow {
	if m.QueryRowFunc != nil {
		return m.QueryRowFunc(ctx, sql, args...)
	}
	return nil
}

func setupMockDB[T any](t *testing.T, mockData []T, scanFunc func(T, []interface{}) error) *mockDB {
	t.Helper()
	return &mockDB{
		QueryFunc: func(ctx context.Context, sql string, args ...interface{}) (interfaces.DBRows, error) {
			rows := &mockRows{
				data:    make([][]interface{}, len(mockData)),
				current: 0,
			}

			for i, data := range mockData {
				rows.data[i] = make([]interface{}, 3)
				if err := scanFunc(data, rows.data[i]); err != nil {
					t.Fatalf("Failed to setup mock data: %v", err)
				}
			}

			rows.scanFunc = func(dest ...interface{}) error {
				if rows.current <= 0 || rows.current > len(rows.data) {
					return fmt.Errorf("invalid row access")
				}
				for i := range dest {
					*(dest[i].(*string)) = rows.data[rows.current-1][i].(string)
				}
				return nil
			}

			return rows, nil
		},
	}
}

func TestPIPHandler_HandleGetRoles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		userID     string
		mockData   []model.Role
		wantStatus int
	}{
		{
			name:   "success_get_role",
			userID: "user1",
			mockData: []model.Role{
				{ID: "1", Name: "admin", Description: "Administrator"},
				{ID: "2", Name: "user", Description: "General User"},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "success_no_role",
			userID:     "user2",
			mockData:   []model.Role{},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockDB := setupMockDB(t, tt.mockData, func(role model.Role, dest []interface{}) error {
				dest[0] = role.ID
				dest[1] = role.Name
				dest[2] = role.Description
				return nil
			})

			handler := NewPIPHandler(mockDB)
			req := httptest.NewRequest("GET", "/users/"+tt.userID+"/roles", nil)
			rec := httptest.NewRecorder()
			handler.HandleGetRoles(rec, req, tt.userID)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %v, want %v", rec.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var roles []model.Role
				if err := json.NewDecoder(rec.Body).Decode(&roles); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if len(roles) != len(tt.mockData) {
					t.Errorf("got %d roles, want %d", len(roles), len(tt.mockData))
				}
			}
		})
	}
}

func setupAttributesMockDB(t *testing.T, mockData map[string]interface{}) *mockDB {
	t.Helper()
	return &mockDB{
		QueryFunc: func(ctx context.Context, sql string, args ...interface{}) (interfaces.DBRows, error) {
			rows := &mockRows{
				data:    make([][]interface{}, len(mockData)),
				current: 0,
			}

			i := 0
			for k, v := range mockData {
				rows.data[i] = []interface{}{k, v.(string)}
				i++
			}

			rows.scanFunc = func(dest ...interface{}) error {
				if rows.current <= 0 || rows.current > len(rows.data) {
					return fmt.Errorf("invalid row access")
				}
				d1 := dest[0].(*string)
				d2 := dest[1].(*string)
				*d1 = rows.data[rows.current-1][0].(string)
				*d2 = rows.data[rows.current-1][1].(string)
				return nil
			}

			return rows, nil
		},
	}
}

func TestPIPHandler_HandleGetAttributes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		userID     string
		mockData   map[string]interface{}
		wantStatus int
	}{
		{
			name:   "success_get_attributes",
			userID: "user1",
			mockData: map[string]interface{}{
				"department": "Engineering",
				"position":   "Developer",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "success_no_attributes",
			userID:     "user2",
			mockData:   map[string]interface{}{},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockDB := setupAttributesMockDB(t, tt.mockData)
			handler := NewPIPHandler(mockDB)
			req := httptest.NewRequest("GET", "/users/"+tt.userID+"/attributes", nil)
			rec := httptest.NewRecorder()
			handler.HandleGetAttributes(rec, req, tt.userID)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %v, want %v", rec.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var attrs map[string]interface{}
				if err := json.NewDecoder(rec.Body).Decode(&attrs); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if len(attrs) != len(tt.mockData) {
					t.Errorf("got %d attributes, want %d", len(attrs), len(tt.mockData))
				}
			}
		})
	}
}

func TestPIPHandler_HandleGetRelationships(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		userID     string
		mockData   []model.Relationship
		wantStatus int
	}{
		{
			name:   "success_get_relationships",
			userID: "user1",
			mockData: []model.Relationship{
				{
					SubjectID: "user1",
					ObjectID:  "resource1",
					Type:      "owner",
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "success_no_relationships",
			userID:     "user2",
			mockData:   []model.Relationship{},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockDB := setupMockDB(t, tt.mockData, func(rel model.Relationship, dest []interface{}) error {
				dest[0] = rel.SubjectID
				dest[1] = rel.ObjectID
				dest[2] = rel.Type
				return nil
			})

			handler := NewPIPHandler(mockDB)
			req := httptest.NewRequest("GET", "/users/"+tt.userID+"/relationships", nil)
			rec := httptest.NewRecorder()
			handler.HandleGetRelationships(rec, req, tt.userID)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %v, want %v", rec.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var relationships []model.Relationship
				if err := json.NewDecoder(rec.Body).Decode(&relationships); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if len(relationships) != len(tt.mockData) {
					t.Errorf("got %d relationships, want %d", len(relationships), len(tt.mockData))
				}
			}
		})
	}
}
