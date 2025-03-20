package interfaces

import (
	"context"
	"net/http"

	"github.com/bmf-san/poc-opa-access-control-system/internal/model"
)

// PolicyEvaluator represents a policy evaluation service
type PolicyEvaluator interface {
	Evaluate(ctx context.Context, req model.EvaluationRequest) (bool, error)
}

// PolicyInformationProvider represents a policy information point service
type PolicyInformationProvider interface {
	GetRoles(ctx context.Context, userID string) ([]model.Role, error)
	GetAttributes(ctx context.Context, userID string) (map[string]interface{}, error)
	GetRelationships(ctx context.Context, userID string) ([]model.Relationship, error)
}

// Repository represents a data access layer
type Repository interface {
	GetUserRoles(ctx context.Context, userID string) ([]string, []model.RBACPermission, error)
	GetUserAttributes(ctx context.Context, userID string) (*model.UserAttributes, error)
	GetResourceAttributes(ctx context.Context, resourceID string) (*model.ResourceAttributes, error)
	GetUserRelationships(ctx context.Context, userID string) ([]model.Relationship, error)
	GetResourceIDByType(ctx context.Context, resourceType string) (string, error)
}

// HTTPClient represents an HTTP client interface
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
