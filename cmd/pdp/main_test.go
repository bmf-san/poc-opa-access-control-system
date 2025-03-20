package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bmf-san/poc-opa-access-control-system/internal/mocks"
	"github.com/bmf-san/poc-opa-access-control-system/internal/model"
)

func TestPDPHandler_HandleEvaluation(t *testing.T) {
	tests := []struct {
		name         string
		request      model.EvaluationRequest
		mockRepo     *mocks.MockRepository
		wantStatus   int
		wantResponse *model.PolicyResponse
	}{
		{
			name: "Manager_permission",
			request: model.EvaluationRequest{
				UserID:       "user1",
				ResourceType: "employees",
				ResourceID:   "11111111-1111-1111-1111-111111111111",
				Action:       "view",
			},
			mockRepo: &mocks.MockRepository{
				GetUserRolesFunc: func(ctx context.Context, userID string) ([]string, []model.RBACPermission, error) {
					return []string{"11111111-1111-1111-1111-111111111111"}, []model.RBACPermission{
						{Role: "11111111-1111-1111-1111-111111111111", ResourceID: "11111111-1111-1111-1111-111111111111", Action: "view"},
					}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantResponse: &model.PolicyResponse{
				Allow:         true,
				Message:       "Access granted",
				AllowedFields: []string{"id", "name", "email", "department_id", "department_name", "employment_type_id", "employment_type", "position", "joined_at"},
			},
		},
		{
			name: "Employee_permission",
			request: model.EvaluationRequest{
				UserID:       "user2",
				ResourceType: "employees",
				ResourceID:   "11111111-1111-1111-1111-111111111111",
				Action:       "view",
			},
			mockRepo: &mocks.MockRepository{
				GetUserRolesFunc: func(ctx context.Context, userID string) ([]string, []model.RBACPermission, error) {
					return []string{"22222222-2222-2222-2222-222222222222"}, []model.RBACPermission{
						{Role: "22222222-2222-2222-2222-222222222222", ResourceID: "11111111-1111-1111-1111-111111111111", Action: "view"},
					}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantResponse: &model.PolicyResponse{
				Allow:         true,
				Message:       "Access granted",
				AllowedFields: []string{"id", "name", "department_name", "employment_type"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewPDPHandler(tt.mockRepo)

			reqBody, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			req := httptest.NewRequest("POST", "/evaluation", bytes.NewBuffer(reqBody))
			rec := httptest.NewRecorder()

			handler.HandleEvaluation(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("HandleEvaluation() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			if tt.wantResponse != nil {
				var resp model.PolicyResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if resp.Allow != tt.wantResponse.Allow {
					t.Errorf("HandleEvaluation() allowed = %v, want %v", resp.Allow, tt.wantResponse.Allow)
				}

				if len(tt.wantResponse.AllowedFields) > 0 {
					if len(resp.AllowedFields) != len(tt.wantResponse.AllowedFields) {
						t.Errorf("HandleEvaluation() allowed fields length = %v, want %v",
							len(resp.AllowedFields), len(tt.wantResponse.AllowedFields))
					}
					for i, field := range tt.wantResponse.AllowedFields {
						if i < len(resp.AllowedFields) && resp.AllowedFields[i] != field {
							t.Errorf("HandleEvaluation() allowed field = %v, want %v",
								resp.AllowedFields[i], field)
						}
					}
				}
			}
		})
	}
}

func TestPDPHandler_evaluateRBAC(t *testing.T) {
	tests := []struct {
		name      string
		request   model.EvaluationRequest
		mockRepo  *mocks.MockRepository
		want      model.PolicyResponse
		wantError bool
	}{
		{
			name: "Manager_role_full_access",
			request: model.EvaluationRequest{
				UserID:       "user1",
				ResourceType: "employees",
				ResourceID:   "11111111-1111-1111-1111-111111111111",
				Action:       "view",
			},
			mockRepo: &mocks.MockRepository{
				GetUserRolesFunc: func(ctx context.Context, userID string) ([]string, []model.RBACPermission, error) {
					return []string{"11111111-1111-1111-1111-111111111111"}, []model.RBACPermission{
						{Role: "11111111-1111-1111-1111-111111111111", ResourceID: "11111111-1111-1111-1111-111111111111", Action: "view"},
					}, nil
				},
			},
			want: model.PolicyResponse{
				Allow:         true,
				Message:       "Access granted",
				AllowedFields: []string{"id", "name", "email", "department_id", "department_name", "employment_type_id", "employment_type", "position", "joined_at"},
			},
			wantError: false,
		},
		{
			name: "Employee_role_restricted_access",
			request: model.EvaluationRequest{
				UserID:       "user2",
				ResourceType: "employees",
				ResourceID:   "11111111-1111-1111-1111-111111111111",
				Action:       "view",
			},
			mockRepo: &mocks.MockRepository{
				GetUserRolesFunc: func(ctx context.Context, userID string) ([]string, []model.RBACPermission, error) {
					return []string{"22222222-2222-2222-2222-222222222222"}, []model.RBACPermission{
						{Role: "22222222-2222-2222-2222-222222222222", ResourceID: "11111111-1111-1111-1111-111111111111", Action: "view"},
					}, nil
				},
			},
			want: model.PolicyResponse{
				Allow:         true,
				Message:       "Access granted",
				AllowedFields: []string{"id", "name", "department_name", "employment_type"},
			},
			wantError: false,
		},
		{
			name: "No_access",
			request: model.EvaluationRequest{
				UserID:       "user3",
				ResourceType: "employees",
				ResourceID:   "11111111-1111-1111-1111-111111111111",
				Action:       "view",
			},
			mockRepo: &mocks.MockRepository{
				GetUserRolesFunc: func(ctx context.Context, userID string) ([]string, []model.RBACPermission, error) {
					return []string{}, []model.RBACPermission{}, nil
				},
			},
			want: model.PolicyResponse{
				Allow:   false,
				Message: "Access denied",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewPDPHandler(tt.mockRepo)
			got, err := handler.evaluateRBAC(context.Background(), tt.request)

			if (err != nil) != tt.wantError {
				t.Errorf("evaluateRBAC() error = %v, wantErr %v", err, tt.wantError)
				return
			}

			if got.Allow != tt.want.Allow {
				t.Errorf("evaluateRBAC() allow = %v, want %v", got.Allow, tt.want.Allow)
			}

			if len(tt.want.AllowedFields) > 0 {
				if len(got.AllowedFields) != len(tt.want.AllowedFields) {
					t.Errorf("evaluateRBAC() allowed fields length = %v, want %v",
						len(got.AllowedFields), len(tt.want.AllowedFields))
				}
				for i, field := range tt.want.AllowedFields {
					if i >= len(got.AllowedFields) {
						t.Errorf("evaluateRBAC() missing field %s", field)
						continue
					}
					if got.AllowedFields[i] != field {
						t.Errorf("evaluateRBAC() allowed field = %v, want %v",
							got.AllowedFields[i], field)
					}
				}
			}
		})
	}
}
