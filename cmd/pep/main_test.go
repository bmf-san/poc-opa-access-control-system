package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/bmf-san/poc-opa-access-control-system/internal/model"
)

// Mock ResourceRepository for testing
type mockResourceRepository struct {
	returnID  string
	returnErr error
}

func (m *mockResourceRepository) GetResourceIDByType(ctx context.Context, resourceType string) (string, error) {
	return m.returnID, m.returnErr
}

func TestProxyHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		userID         string
		policyType     string
		mockPDPResp    *model.PolicyResponse
		mockResourceID string
		wantStatusCode int
	}{
		{
			name:       "manager_role_full_field_access",
			method:     "GET",
			path:       "/employees",
			userID:     "user1",
			policyType: "rbac",
			mockPDPResp: &model.PolicyResponse{
				Allow:         true,
				Message:       "Access granted",
				AllowedFields: []string{"id", "name", "email", "department_id", "department_name", "position", "joined_at"},
				FilteredData: map[string][]map[string]interface{}{
					"employees": {
						{
							"id":              "11111111-1111-1111-1111-111111111111",
							"name":            "John Doe",
							"email":           "john.doe@example.com",
							"department_id":   "dep1",
							"department_name": "Engineering",
							"position":        "Engineer",
							"joined_at":       "2023-01-01T00:00:00Z",
						},
					},
				},
			},
			mockResourceID: "11111111-1111-1111-1111-111111111111", // employees resource
			wantStatusCode: http.StatusOK,
		},
		{
			name:       "employee_role_restricted_access",
			method:     "GET",
			path:       "/employees",
			userID:     "user2",
			policyType: "rbac",
			mockPDPResp: &model.PolicyResponse{
				Allow:         true,
				Message:       "Access granted",
				AllowedFields: []string{"id", "name"},
				FilteredData: map[string][]map[string]interface{}{
					"employees": {
						{
							"id":   "11111111-1111-1111-1111-111111111111",
							"name": "John Doe",
						},
					},
				},
			},
			mockResourceID: "11111111-1111-1111-1111-111111111111", // employees resource
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "missing_x_user_id_header",
			method:         "GET",
			path:           "/employees",
			userID:         "",
			policyType:     "rbac",
			mockResourceID: "11111111-1111-1111-1111-111111111111", // employees resource
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "non_resource_path",
			method:         "GET",
			path:           "/health",
			userID:         "user1",
			policyType:     "rbac",
			mockResourceID: "health",
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test server to mock PDP responses
			pdpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if tt.mockPDPResp != nil {
					if !tt.mockPDPResp.Allow {
						w.WriteHeader(http.StatusForbidden)
					}
					json.NewEncoder(w).Encode(tt.mockPDPResp)
				}
			}))
			defer pdpServer.Close()

			// Setup target server
			targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Path == "/employees" {
					response := map[string][]map[string]interface{}{
						"employees": {
							{
								"id":              "11111111-1111-1111-1111-111111111111",
								"name":            "John Doe",
								"email":           "john.doe@example.com",
								"department_id":   "dep1",
								"department_name": "Engineering",
								"position":        "Engineer",
								"joined_at":       "2023-01-01T00:00:00Z",
							},
						},
					}
					json.NewEncoder(w).Encode(response)
				} else {
					w.WriteHeader(http.StatusOK)
				}
			}))
			defer targetServer.Close()

			mockRepo := &mockResourceRepository{returnID: tt.mockResourceID}
			handler := NewProxyHandler(pdpServer.URL, mockRepo)

			// Set custom director for test
			handler.SetDirector(func(req *http.Request) {
				targetURL, _ := url.Parse(targetServer.URL)
				req.URL.Scheme = targetURL.Scheme
				req.URL.Host = targetURL.Host
				req.Host = targetURL.Host
			})

			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.userID != "" {
				req.Header.Set("X-User-ID", tt.userID)
			}
			if tt.policyType != "" {
				req.Header.Set("X-Policy-Type", tt.policyType)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatusCode {
				t.Errorf("Status code = %v, want %v", rec.Code, tt.wantStatusCode)
			}

			// Verify response content
			if tt.mockPDPResp != nil && tt.mockPDPResp.Allow && tt.path == "/employees" {
				var response map[string][]map[string]interface{}
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				employees := response["employees"]
				if len(employees) != 1 {
					t.Errorf("Expected 1 employee, got %d", len(employees))
					return
				}

				emp := employees[0]
				allowedFields := tt.mockPDPResp.AllowedFields
				for _, field := range allowedFields {
					if _, exists := emp[field]; !exists {
						t.Errorf("Expected field %s in response", field)
					}
				}

				// Verify unauthorized fields are not included in the response
				allFields := []string{"id", "name", "email", "department_id", "department_name", "position", "joined_at"}
				for _, field := range allFields {
					if _, exists := emp[field]; exists {
						found := false
						for _, allowed := range allowedFields {
							if field == allowed {
								found = true
								break
							}
						}
						if !found {
							t.Errorf("Unauthorized field %s found in response", field)
						}
					}
				}
			}
		})
	}
}

func TestProxyHandler_checkAccess(t *testing.T) {
	tests := []struct {
		name         string
		userID       string
		resourceType string
		resourceID   string
		action       string
		policyType   string
		mockPDPResp  *model.PolicyResponse
		want         model.PolicyResponse
		wantErr      bool
	}{
		{
			name:         "manager_role_allowed_access",
			userID:       "11111111-1111-1111-1111-111111111111",
			resourceType: "employees",
			resourceID:   "11111111-1111-1111-1111-111111111111",
			action:       "view",
			policyType:   "rbac",
			mockPDPResp: &model.PolicyResponse{
				Allow:         true,
				Message:       "Access granted",
				AllowedFields: []string{"id", "name", "email", "department_id", "department_name", "position", "joined_at"},
				FilteredData: map[string][]map[string]interface{}{
					"employees": {
						{
							"id":              "11111111-1111-1111-1111-111111111111",
							"name":            "John Doe",
							"email":           "john.doe@example.com",
							"department_id":   "dep1",
							"department_name": "Engineering",
							"position":        "Engineer",
							"joined_at":       "2023-01-01T00:00:00Z",
						},
					},
				},
			},
			want: model.PolicyResponse{
				Allow:         true,
				Message:       "Access granted",
				AllowedFields: []string{"id", "name", "email", "department_id", "department_name", "position", "joined_at"},
				FilteredData: map[string][]map[string]interface{}{
					"employees": {
						{
							"id":              "11111111-1111-1111-1111-111111111111",
							"name":            "John Doe",
							"email":           "john.doe@example.com",
							"department_id":   "dep1",
							"department_name": "Engineering",
							"position":        "Engineer",
							"joined_at":       "2023-01-01T00:00:00Z",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:         "employee_role_restricted_access",
			userID:       "22222222-2222-2222-2222-222222222222",
			resourceType: "employees",
			resourceID:   "11111111-1111-1111-1111-111111111111",
			action:       "view",
			policyType:   "rbac",
			mockPDPResp: &model.PolicyResponse{
				Allow:         true,
				Message:       "Access granted",
				AllowedFields: []string{"id", "name"},
				FilteredData: map[string][]map[string]interface{}{
					"employees": {
						{
							"id":   "11111111-1111-1111-1111-111111111111",
							"name": "John Doe",
						},
					},
				},
			},
			want: model.PolicyResponse{
				Allow:         true,
				Message:       "Access granted",
				AllowedFields: []string{"id", "name"},
				FilteredData: map[string][]map[string]interface{}{
					"employees": {
						{
							"id":   "11111111-1111-1111-1111-111111111111",
							"name": "John Doe",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:         "policy_type_not_specified",
			userID:       "11111111-1111-1111-1111-111111111111",
			resourceType: "employees",
			resourceID:   "11111111-1111-1111-1111-111111111111",
			action:       "view",
			policyType:   "",
			mockPDPResp:  nil,
			want:         model.PolicyResponse{},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pdpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if tt.mockPDPResp != nil {
					if !tt.mockPDPResp.Allow {
						w.WriteHeader(http.StatusForbidden)
					}
					json.NewEncoder(w).Encode(tt.mockPDPResp)
				}
			}))
			defer pdpServer.Close()

			mockRepo := &mockResourceRepository{returnID: tt.resourceID}
			handler := NewProxyHandler(pdpServer.URL, mockRepo)

			// Set custom director for test
			handler.SetDirector(func(req *http.Request) {
				// No need to modify requests since this test doesn't actually proxy requests
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("X-Policy-Type", tt.policyType)

			got, err := handler.checkAccess(req, model.EvaluationRequest{
				UserID:       tt.userID,
				ResourceType: tt.resourceType,
				ResourceID:   tt.resourceID,
				Action:       tt.action,
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("checkAccess() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Allow != tt.want.Allow {
				t.Errorf("checkAccess() Allow = %v, want %v", got.Allow, tt.want.Allow)
			}
			if !tt.wantErr {
				if len(got.AllowedFields) != len(tt.want.AllowedFields) {
					t.Errorf("checkAccess() AllowedFields length = %v, want %v", len(got.AllowedFields), len(tt.want.AllowedFields))
				}
				if tt.want.FilteredData != nil {
					if got.FilteredData == nil {
						t.Error("checkAccess() FilteredData is nil, but expected data")
					}
				}
			}
		})
	}
}
