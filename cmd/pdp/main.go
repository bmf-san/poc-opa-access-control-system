package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/open-policy-agent/opa/rego"

	"github.com/bmf-san/poc-opa-access-control-system/internal/interfaces"
	"github.com/bmf-san/poc-opa-access-control-system/internal/model"
	"github.com/bmf-san/poc-opa-access-control-system/internal/pkg"
	"github.com/bmf-san/poc-opa-access-control-system/internal/repository"
)

func main() {
	log.Printf("[INFO] Starting PDP server on :8081")

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
	db, err := dbManager.GetClient("prp")
	if err != nil {
		log.Fatalf("[ERROR] Failed to connect to database: %v", err)
	}

	log.Printf("[INFO] Successfully connected to database")

	repo := repository.NewRepository(db)

	// Initialize PDP handler
	pdpHandler := NewPDPHandler(repo)

	// Set up routing
	mux := http.NewServeMux()
	mux.HandleFunc("/evaluation", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		pdpHandler.HandleEvaluation(w, r)
	})

	server := &http.Server{
		Handler:      mux,
		Addr:         "0.0.0.0:8081",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}

// PDPHandler handles PDP requests
type PDPHandler struct {
	repo    interfaces.Repository
	opaRBAC *rego.PreparedEvalQuery
}

// NewPDPHandler creates a new PDPHandler
func NewPDPHandler(repo interfaces.Repository) *PDPHandler {
	// Load policy files
	rbacPolicy, err := loadPolicy("policy/rbac.rego")
	if err != nil {
		log.Fatalf("[ERROR] Failed to load RBAC policy: %v", err)
	}

	// Prepare OPA queries
	opaRBAC, err := rego.New(
		rego.Query("data.policy.rbac.result"),
		rego.Module("rbac.rego", rbacPolicy),
	).PrepareForEval(context.Background())
	if err != nil {
		log.Fatalf("[ERROR] Failed to prepare RBAC policy: %v", err)
	}

	return &PDPHandler{
		repo:    repo,
		opaRBAC: &opaRBAC,
	}
}

func loadPolicy(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read policy file: %w", err)
	}

	return string(content), nil
}

// HandleEvaluation handles policy evaluation requests
func (h *PDPHandler) HandleEvaluation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.Printf("[INFO] Received evaluation request")

	var req model.EvaluationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] Error decoding request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[DEBUG] Processing evaluation request: %+v", req)
	if req.Data != nil {
		log.Printf("[DEBUG] Request includes data for filtering: %+v", req.Data)
	}

	response, err := h.evaluateRBAC(ctx, req)
	if err != nil {
		log.Printf("[ERROR] Evaluation error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare detailed log message
	logMsg := fmt.Sprintf("[INFO] Access Decision:\n"+
		"- Allow: %v\n"+
		"- User ID: %s\n"+
		"- Resource Type: %s\n"+
		"- Resource ID: %s\n"+
		"- Action: %s\n"+
		"- Allowed Fields: %v\n"+
		"- Filtered Data Present: %v",
		response.Allow,
		req.UserID,
		req.ResourceType,
		req.ResourceID,
		req.Action,
		response.AllowedFields,
		response.FilteredData != nil)

	log.Print(logMsg)

	w.Header().Set("Content-Type", "application/json")
	if !response.Allow {
		w.WriteHeader(http.StatusForbidden)
	}
	json.NewEncoder(w).Encode(response)
}

func (h *PDPHandler) evaluateRBAC(ctx context.Context, req model.EvaluationRequest) (model.PolicyResponse, error) {
	log.Printf("[DEBUG] Starting RBAC evaluation for user %s", req.UserID)

	// Get user roles and permissions from repository
	roles, permissions, err := h.repo.GetUserRoles(ctx, req.UserID)
	if err != nil {
		return model.PolicyResponse{}, err
	}

	log.Printf("[DEBUG] User roles: %v", roles)
	log.Printf("[DEBUG] User permissions: %v", permissions)

	// Convert roles and permissions to maps
	userRoles := make([]map[string]interface{}, len(roles))
	for i, role := range roles {
		userRoles[i] = map[string]interface{}{
			"user_id": req.UserID,
			"role_id": role,
		}
	}

	rolePermissions := make([]map[string]interface{}, len(permissions))
	for i, perm := range permissions {
		rolePermissions[i] = map[string]interface{}{
			"role_id":     perm.Role,
			"resource_id": perm.ResourceID,
			"action_id":   perm.Action,
		}
	}

	// Prepare input for OPA
	input := map[string]interface{}{
		"user": map[string]interface{}{
			"id": req.UserID,
		},
		"user_roles":       userRoles,
		"role_permissions": rolePermissions,
		"resource": map[string]interface{}{
			"id":   req.ResourceID,
			"name": req.ResourceType,
		},
		"action": map[string]interface{}{
			"id":   req.Action,
			"name": req.Action,
		},
	}

	// Add data if present
	if req.Data != nil {
		input["data"] = req.Data
		log.Printf("[DEBUG] Including data in policy evaluation: %+v", req.Data)
	}

	log.Printf("[DEBUG] Policy input: %+v", input)

	// Evaluate policy
	results, err := h.opaRBAC.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return model.PolicyResponse{}, fmt.Errorf("policy evaluation error: %w", err)
	}

	// Check if we have any results
	if len(results) == 0 {
		log.Printf("[DEBUG] No policy results")
		return model.PolicyResponse{
			Allow:   false,
			Message: "Access denied - no policy result",
		}, nil
	}

	// Get evaluation result
	resultValue := results[0].Expressions[0].Value
	if resultValue == nil {
		log.Printf("[DEBUG] Empty policy result")
		return model.PolicyResponse{
			Allow:   false,
			Message: "Access denied - empty policy result",
		}, nil
	}

	// Try to convert the result to a map
	result, ok := resultValue.(map[string]interface{})
	if !ok {
		log.Printf("[ERROR] Invalid policy result format: %T", resultValue)
		return model.PolicyResponse{}, fmt.Errorf("invalid policy result format: expected map, got %T", resultValue)
	}

	log.Printf("[DEBUG] Raw policy result: %+v", result)

	// Get the allow value and allowed fields
	allowed, _ := result["allow"].(bool)
	allowedFieldsRaw, _ := result["allowed_fields"].([]interface{})
	allowedFields := make([]string, len(allowedFieldsRaw))
	for i, field := range allowedFieldsRaw {
		if str, ok := field.(string); ok {
			allowedFields[i] = str
		}
	}

	// Get filtered data if present
	var filteredData interface{}
	if result["filtered_data"] != nil {
		filteredData = result["filtered_data"]
		log.Printf("[DEBUG] Filtered data present: %+v", filteredData)
	}

	response := model.PolicyResponse{
		Allow:         allowed,
		Message:       fmt.Sprintf("Access %s", map[bool]string{true: "granted", false: "denied"}[allowed]),
		AllowedFields: allowedFields,
		FilteredData:  filteredData,
	}

	log.Printf("[INFO] Final policy response: %+v", response)
	return response, nil
}
