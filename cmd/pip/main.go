package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/bmf-san/poc-opa-access-control-system/internal/interfaces"
	"github.com/bmf-san/poc-opa-access-control-system/internal/model"
	"github.com/bmf-san/poc-opa-access-control-system/internal/pkg"
)

// DBConnWrapper wraps a pgx.Conn to implement interfaces.DBConn
type DBConnWrapper struct {
	conn *pgx.Conn
}

// Query implements interfaces.DBConn
func (w *DBConnWrapper) Query(ctx context.Context, sql string, args ...interface{}) (interfaces.DBRows, error) {
	rows, err := w.conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &DBRowsWrapper{rows: rows}, nil
}

// QueryRow implements interfaces.DBConn
func (w *DBConnWrapper) QueryRow(ctx context.Context, sql string, args ...interface{}) interfaces.DBRow {
	return &DBRowWrapper{row: w.conn.QueryRow(ctx, sql, args...)}
}

// DBRowsWrapper wraps pgx.Rows to implement interfaces.DBRows
type DBRowsWrapper struct {
	rows pgx.Rows
}

func (w *DBRowsWrapper) Close() {
	w.rows.Close()
}

func (w *DBRowsWrapper) Err() error {
	return w.rows.Err()
}

func (w *DBRowsWrapper) Next() bool {
	return w.rows.Next()
}

func (w *DBRowsWrapper) Scan(dest ...interface{}) error {
	return w.rows.Scan(dest...)
}

// DBRowWrapper wraps pgx.Row to implement interfaces.DBRow
type DBRowWrapper struct {
	row pgx.Row
}

func (w *DBRowWrapper) Scan(dest ...interface{}) error {
	return w.row.Scan(dest...)
}

func main() {
	log.Printf("Starting PIP server on :8082")

	// Initialize database manager
	dbConfig := pkg.DBConfig{
		Host:     "prp-db",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "prp",
		SSLMode:  "disable",
	}

	dbManager := pkg.NewDBManager(map[string]pkg.DBConfig{
		"prp": dbConfig,
	})
	defer dbManager.CloseAll()

	// Get database client
	dbConn, err := dbManager.GetClient("prp")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Printf("Successfully connected to database")

	// Initialize PIP handler with wrapped connection
	pipHandler := NewPIPHandler(&DBConnWrapper{conn: dbConn})

	// Set up routing
	mux := http.NewServeMux()

	// Handler functions to route requests based on path pattern
	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse path to extract user_id and endpoint
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) != 3 {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		userID := parts[1]
		endpoint := parts[2]

		switch endpoint {
		case "roles":
			pipHandler.HandleGetRoles(w, r, userID)
		case "attributes":
			pipHandler.HandleGetAttributes(w, r, userID)
		case "relationships":
			pipHandler.HandleGetRelationships(w, r, userID)
		default:
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})

	server := &http.Server{
		Handler:      mux,
		Addr:         "0.0.0.0:8082",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}

// PIPHandler handles Policy Information Point requests
type PIPHandler struct {
	db interfaces.DBConn
}

// NewPIPHandler creates a new PIPHandler
func NewPIPHandler(db interfaces.DBConn) *PIPHandler {
	return &PIPHandler{db: db}
}

// HandleGetRoles retrieves user roles
func (h *PIPHandler) HandleGetRoles(w http.ResponseWriter, r *http.Request, userID string) {
	ctx := r.Context()

	rows, err := h.db.Query(ctx, `
SELECT r.id, r.name, r.description
FROM roles r
JOIN user_roles ur ON r.id = ur.role_id
WHERE ur.user_id = $1
`, userID)
	if err != nil {
		log.Printf("Database error querying roles: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	roles := make([]model.Role, 0)
	for rows.Next() {
		var role model.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description); err != nil {
			log.Printf("Error scanning role: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		roles = append(roles, role)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating roles: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roles)
}

// HandleGetRelationships retrieves user's relationships
func (h *PIPHandler) HandleGetRelationships(w http.ResponseWriter, r *http.Request, userID string) {
	ctx := r.Context()

	rows, err := h.db.Query(ctx, `
        SELECT r.subject_id, r.object_id, r.relationship_type
        FROM relationships r
        WHERE r.subject_id = $1 OR r.object_id = $1
    `, userID)
	if err != nil {
		log.Printf("Database error querying relationships: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	relationships := make([]model.Relationship, 0)
	for rows.Next() {
		var rel model.Relationship
		if err := rows.Scan(&rel.SubjectID, &rel.ObjectID, &rel.Type); err != nil {
			log.Printf("Error scanning relationship: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		relationships = append(relationships, rel)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating relationships: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(relationships)
}

// HandleGetAttributes retrieves user attributes
func (h *PIPHandler) HandleGetAttributes(w http.ResponseWriter, r *http.Request, userID string) {
	ctx := r.Context()

	rows, err := h.db.Query(ctx, `
SELECT attr.name, attr.value
FROM attributes attr
WHERE attr.user_id = $1
`, userID)
	if err != nil {
		log.Printf("Database error querying attributes: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	attributes := make(map[string]interface{})
	for rows.Next() {
		var name, value string
		if err := rows.Scan(&name, &value); err != nil {
			log.Printf("Error scanning attribute: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		attributes[name] = value
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating attributes: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(attributes)
}
