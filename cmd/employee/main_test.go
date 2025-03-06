package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// mockDB implements the DB interface for testing
type mockDB struct {
	queryRowFunc func(ctx context.Context, sql string, args ...interface{}) pgx.Row
	queryFunc    func(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

func (m *mockDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return m.queryRowFunc(ctx, sql, args...)
}

func (m *mockDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return m.queryFunc(ctx, sql, args...)
}

// mockRows implements pgx.Rows for testing
type mockRows struct {
	employees []Employee
	current   int
	err       error
}

func (m *mockRows) Close()                                       {}
func (m *mockRows) Err() error                                   { return m.err }
func (m *mockRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (m *mockRows) FieldDescriptions() []pgconn.FieldDescription { return []pgconn.FieldDescription{} }
func (m *mockRows) Conn() *pgx.Conn                              { return nil }
func (m *mockRows) RawValues() [][]byte                          { return nil }
func (m *mockRows) Values() ([]interface{}, error)               { return nil, nil }
func (m *mockRows) Next() bool {
	m.current++
	return m.current <= len(m.employees)
}
func (m *mockRows) Scan(dest ...interface{}) error {
	if m.current > len(m.employees) {
		return pgx.ErrNoRows
	}
	emp := m.employees[m.current-1]
	*dest[0].(*string) = emp.ID
	*dest[1].(*string) = emp.Name
	*dest[2].(*string) = emp.Email
	*dest[3].(*string) = emp.DepartmentID
	*dest[4].(*string) = emp.DepartmentName
	*dest[5].(*string) = emp.EmploymentTypeID
	*dest[6].(*string) = emp.EmploymentType
	*dest[7].(*string) = emp.Position
	*dest[8].(*time.Time) = emp.JoinedAt
	return nil
}

func TestHandleListEmployees(t *testing.T) {
	// Test data
	joinedAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	testEmployees := []Employee{
		{
			ID:               "emp1",
			Name:             "John Doe",
			Email:            "john.doe@example.com",
			DepartmentID:     "dep1",
			DepartmentName:   "Engineering",
			EmploymentTypeID: "type1",
			EmploymentType:   "Full-time",
			Position:         "Software Engineer",
			JoinedAt:         joinedAt,
		},
		{
			ID:               "emp2",
			Name:             "Jane Smith",
			Email:            "jane.smith@example.com",
			DepartmentID:     "dep2",
			DepartmentName:   "Design",
			EmploymentTypeID: "type1",
			EmploymentType:   "Full-time",
			Position:         "UI Designer",
			JoinedAt:         joinedAt,
		},
	}

	tests := []struct {
		name       string
		mockDB     *mockDB
		wantStatus int
		wantLen    int
	}{
		{
			name: "success_get_employee_list",
			mockDB: &mockDB{
				queryFunc: func(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
					return &mockRows{employees: testEmployees}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantLen:    2,
		},
		{
			name: "success_no_employees",
			mockDB: &mockDB{
				queryFunc: func(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
					return &mockRows{employees: []Employee{}}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantLen:    0,
		},
		{
			name: "error_database_error",
			mockDB: &mockDB{
				queryFunc: func(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
					return nil, pgx.ErrTxClosed
				},
			},
			wantStatus: http.StatusInternalServerError,
			wantLen:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewEmployeeHandler(tt.mockDB)
			mux := http.NewServeMux()
			mux.HandleFunc("/employees", func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
					return
				}
				handler.HandleListEmployees(w, r)
			})

			req := httptest.NewRequest("GET", "/employees", nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("HandleListEmployees() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var response map[string][]Employee
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if len(response["employees"]) != tt.wantLen {
					t.Errorf("HandleListEmployees() returned %d employees, want %d", len(response["employees"]), tt.wantLen)
				}
			}
		})
	}
}
