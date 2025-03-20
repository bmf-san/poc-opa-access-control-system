package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"

	"github.com/bmf-san/poc-opa-access-control-system/internal/interfaces"
	"github.com/bmf-san/poc-opa-access-control-system/internal/model"
)

type Repository struct {
	db *pgx.Conn
}

func NewRepository(db *pgx.Conn) interfaces.Repository {
	return &Repository{db: db}
}

func (r *Repository) GetUserRoles(ctx context.Context, userID string) ([]string, []model.RBACPermission, error) {
	var roles []string
	var permissions []model.RBACPermission

	// Get roles
	rows, err := r.db.Query(ctx, `
		SELECT r.id
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
	`, userID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, nil, err
		}
		roles = append(roles, role)
	}

	// Get permissions
	rows, err = r.db.Query(ctx, `
		SELECT r.id, res.id, a.name
		FROM roles r
		JOIN role_permissions rp ON r.id = rp.role_id
		JOIN resources res ON rp.resource_id = res.id
		JOIN actions a ON rp.action_id = a.id
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
	`, userID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var perm model.RBACPermission
		if err := rows.Scan(&perm.Role, &perm.ResourceID, &perm.Action); err != nil {
			return nil, nil, err
		}
		permissions = append(permissions, perm)
	}

	return roles, permissions, nil
}

func (r *Repository) GetUserAttributes(ctx context.Context, userID string) (*model.UserAttributes, error) {
	user := &model.UserAttributes{}
	err := r.db.QueryRow(ctx, `
		SELECT u.department_id, d.name, u.employment_type_id
		FROM users u
		JOIN departments d ON u.department_id = d.id
		WHERE u.id = $1
	`, userID).Scan(&user.DepartmentID, &user.DepartmentName, &user.EmploymentTypeID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) GetResourceAttributes(ctx context.Context, resourceID string) (*model.ResourceAttributes, error) {
	resource := &model.ResourceAttributes{}
	err := r.db.QueryRow(ctx, `
		SELECT department_id
		FROM users
		WHERE id = $1
	`, resourceID).Scan(&resource.DepartmentID)
	if err != nil {
		return nil, err
	}
	return resource, nil
}

func (r *Repository) GetUserRelationships(ctx context.Context, userID string) ([]model.Relationship, error) {
	var relationships []model.Relationship

	rows, err := r.db.Query(ctx, `
SELECT subject_id, object_id, relation
FROM relationships
WHERE subject_id = $1 OR object_id = $1
`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var rel model.Relationship
		if err := rows.Scan(&rel.SubjectID, &rel.ObjectID, &rel.Type); err != nil {
			return nil, err
		}
		relationships = append(relationships, rel)
	}

	return relationships, nil
}

func (r *Repository) GetResourceIDByType(ctx context.Context, resourceType string) (string, error) {
	log.Printf("[DEBUG] Executing query to get resource ID for type '%s'", resourceType)

	var resourceID string
	err := r.db.QueryRow(ctx, `
SELECT id
FROM resources
WHERE name = $1
  AND tenant_id = '11111111-1111-1111-1111-111111111111'
LIMIT 1
`, resourceType).Scan(&resourceID)

	if err == pgx.ErrNoRows {
		log.Printf("[ERROR] No resource found with type '%s'", resourceType)
		return "", fmt.Errorf("resource type '%s' not found", resourceType)
	}
	if err != nil {
		log.Printf("[ERROR] Database error while getting resource ID: %v", err)
		return "", fmt.Errorf("database error: %v", err)
	}

	log.Printf("[DEBUG] Found resource ID '%s' for type '%s'", resourceID, resourceType)
	return resourceID, nil
}
