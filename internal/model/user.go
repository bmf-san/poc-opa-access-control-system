package model

// User represents a user in the system
type User struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	Department     string `json:"department"`
	EmploymentType string `json:"employment_type"`
	Tenant         string `json:"tenant"`
}

// Role represents a role in the system
type Role struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Resource represents a resource in the system
type Resource struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

// Attribute represents a user attribute
type Attribute struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}
