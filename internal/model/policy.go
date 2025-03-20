package model

type PolicyResponse struct {
	Allow         bool        `json:"allow"`
	Message       string      `json:"message,omitempty"`
	AllowedFields []string    `json:"allowed_fields,omitempty"`
	FilteredData  interface{} `json:"filtered_data,omitempty"`
}

type EvaluationRequest struct {
	UserID       string      `json:"user_id"`
	ResourceType string      `json:"resource_type"`
	ResourceID   string      `json:"resource_id"`
	Action       string      `json:"action"`
	Data         interface{} `json:"data,omitempty"`
}

// RBAC specific types
type RBACPermission struct {
	Role       string `json:"role"`
	ResourceID string `json:"resource_id"`
	Action     string `json:"action"`
}

type RBACPolicy struct {
	Permissions []RBACPermission `json:"permissions"`
}

// UserAttributes represents attributes associated with a user
type UserAttributes struct {
	DepartmentID     string `json:"department_id"`
	DepartmentName   string `json:"department_name"`
	EmploymentTypeID string `json:"employment_type_id"`
}

// ResourceAttributes represents attributes associated with a resource
type ResourceAttributes struct {
	DepartmentID string `json:"department_id"`
}

// Relationship represents a connection between a subject and an object
type Relationship struct {
	SubjectID string `json:"subject_id"`
	ObjectID  string `json:"object_id"`
	Type      string `json:"type"`
}
