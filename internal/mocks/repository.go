package mocks

import (
	"context"

	"github.com/bmf-san/poc-opa-access-control-system/internal/model"
)

// MockRepository is a mock implementation of the Repository interface
type MockRepository struct {
	GetUserRolesFunc          func(ctx context.Context, userID string) ([]string, []model.RBACPermission, error)
	GetUserAttributesFunc     func(ctx context.Context, userID string) (*model.UserAttributes, error)
	GetResourceAttributesFunc func(ctx context.Context, resourceID string) (*model.ResourceAttributes, error)
	GetUserRelationshipsFunc  func(ctx context.Context, userID string) ([]model.Relationship, error)
	GetResourceIDByTypeFunc   func(ctx context.Context, resourceType string) (string, error)
}

func (m *MockRepository) GetUserRoles(ctx context.Context, userID string) ([]string, []model.RBACPermission, error) {
	if m.GetUserRolesFunc != nil {
		return m.GetUserRolesFunc(ctx, userID)
	}
	return nil, nil, nil
}

func (m *MockRepository) GetUserAttributes(ctx context.Context, userID string) (*model.UserAttributes, error) {
	if m.GetUserAttributesFunc != nil {
		return m.GetUserAttributesFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockRepository) GetResourceAttributes(ctx context.Context, resourceID string) (*model.ResourceAttributes, error) {
	if m.GetResourceAttributesFunc != nil {
		return m.GetResourceAttributesFunc(ctx, resourceID)
	}
	return nil, nil
}

func (m *MockRepository) GetUserRelationships(ctx context.Context, userID string) ([]model.Relationship, error) {
	if m.GetUserRelationshipsFunc != nil {
		return m.GetUserRelationshipsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockRepository) GetResourceIDByType(ctx context.Context, resourceType string) (string, error) {
	if m.GetResourceIDByTypeFunc != nil {
		return m.GetResourceIDByTypeFunc(ctx, resourceType)
	}
	return "", nil
}
