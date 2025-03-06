package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/bmf-san/poc-opa-access-control-system/internal/pkg"
)

type Employee struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Email            string    `json:"email"`
	DepartmentID     string    `json:"department_id"`
	DepartmentName   string    `json:"department_name"`
	EmploymentTypeID string    `json:"employment_type_id"`
	EmploymentType   string    `json:"employment_type"`
	Position         string    `json:"position"`
	JoinedAt         time.Time `json:"joined_at"`
}

// DB interface to make testing easier
type DB interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

type EmployeeHandler struct {
	db DB
}

func NewEmployeeHandler(db DB) *EmployeeHandler {
	return &EmployeeHandler{
		db: db,
	}
}

func main() {
	log.Printf("Starting Employee service on :8083")

	// Initialize database manager
	dbConfig := pkg.DBConfig{
		Host:     "employee-db",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "employee",
		SSLMode:  "disable",
	}

	dbManager := pkg.NewDBManager(map[string]pkg.DBConfig{
		"employee": dbConfig,
	})
	defer dbManager.CloseAll()

	// Get database client
	db, err := dbManager.GetClient("employee")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Printf("Successfully connected to database")

	// Initialize handler
	employeeHandler := NewEmployeeHandler(db)

	// Set up routing
	mux := http.NewServeMux()
	mux.HandleFunc("/employees", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		employeeHandler.HandleListEmployees(w, r)
	})

	server := &http.Server{
		Handler:      mux,
		Addr:         "0.0.0.0:8083",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}

func (h *EmployeeHandler) HandleListEmployees(w http.ResponseWriter, r *http.Request) {
	employees, err := h.getAllEmployees(r.Context())
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string][]Employee{
		"employees": employees,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *EmployeeHandler) getAllEmployees(ctx context.Context) ([]Employee, error) {
	query := `
    SELECT
        e.id,
        e.name,
        e.email,
        e.department_id,
        d.name as department_name,
        e.employment_type_id,
        et.name as employment_type,
        e.position,
        e.joined_at
    FROM employees e
    JOIN departments d ON e.department_id = d.id
    JOIN employment_types et ON e.employment_type_id = et.id
    ORDER BY e.name
`

	rows, err := h.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []Employee
	for rows.Next() {
		var emp Employee
		err := rows.Scan(
			&emp.ID,
			&emp.Name,
			&emp.Email,
			&emp.DepartmentID,
			&emp.DepartmentName,
			&emp.EmploymentTypeID,
			&emp.EmploymentType,
			&emp.Position,
			&emp.JoinedAt,
		)
		if err != nil {
			return nil, err
		}
		employees = append(employees, emp)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return employees, nil
}
