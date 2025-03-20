package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/bmf-san/poc-opa-access-control-system/internal/model"
	"github.com/bmf-san/poc-opa-access-control-system/internal/repository"
)

// ResourceRepository defines the interface for resource operations
type ResourceRepository interface {
	GetResourceIDByType(ctx context.Context, resourceType string) (string, error)
}

type ProxyHandler struct {
	proxy        *httputil.ReverseProxy
	pdpHost      string
	resourceRepo ResourceRepository
	director     func(*http.Request)
}

func defaultDirector(req *http.Request) {
	targetHost := req.Host
	if strings.Contains(targetHost, ":") {
		targetHost = strings.Split(targetHost, ":")[0]
	}

	// Convert domain name to container name (e.g., employee.local -> employee)
	if strings.HasSuffix(targetHost, ".local") {
		targetHost = strings.TrimSuffix(targetHost, ".local")
	}

	backendURL := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:8083", targetHost),
	}

	req.URL.Scheme = backendURL.Scheme
	req.URL.Host = backendURL.Host
	req.Host = backendURL.Host

	log.Printf("[DEBUG] Proxying request to %s", backendURL.String())
}

func NewProxyHandler(pdpHost string, resourceRepo ResourceRepository) *ProxyHandler {
	h := &ProxyHandler{
		pdpHost:      pdpHost,
		resourceRepo: resourceRepo,
		director:     defaultDirector,
	}

	h.proxy = &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			h.director(req)
		},
	}

	return h
}

// SetDirector allows overriding the default director function (useful for testing)
func (h *ProxyHandler) SetDirector(director func(*http.Request)) {
	h.director = director
}

type PolicyResponse = model.PolicyResponse

type responseInterceptor struct {
	writer     http.ResponseWriter
	body       []byte
	header     http.Header
	hasContent bool
	statusCode int
}

func (ri *responseInterceptor) Header() http.Header {
	if ri.header == nil {
		ri.header = make(http.Header)
	}
	return ri.header
}

func (ri *responseInterceptor) Write(p []byte) (n int, err error) {
	ri.body = append(ri.body, p...)
	ri.hasContent = true
	return len(p), nil
}

func (ri *responseInterceptor) WriteHeader(statusCode int) {
	ri.statusCode = statusCode
}

func (ri *responseInterceptor) copyHeadersTo(w http.ResponseWriter) {
	for key, values := range ri.header {
		w.Header()[key] = values
	}
}

func (h *ProxyHandler) checkAccess(_ *http.Request, req model.EvaluationRequest) (PolicyResponse, error) {
	log.Printf("[INFO] Checking access with request: %+v", req)

	client := &http.Client{}

	url := fmt.Sprintf("%s/evaluation", h.pdpHost)
	jsonData, err := json.Marshal(req)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal request: %v", err)
		return PolicyResponse{}, err
	}

	httpReq, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		log.Printf("[ERROR] Failed to create request: %v", err)
		return PolicyResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("[ERROR] Failed to send request: %v", err)
		return PolicyResponse{}, err
	}
	defer resp.Body.Close()

	var policyResp PolicyResponse
	if err := json.NewDecoder(resp.Body).Decode(&policyResp); err != nil {
		log.Printf("[ERROR] Failed to decode response: %v", err)
		return PolicyResponse{}, err
	}

	log.Printf("[INFO] Policy evaluation result: allowed=%v", policyResp.Allow)
	return policyResp, nil
}

func (h *ProxyHandler) getResourceID(ctx context.Context, resourceType string) (string, error) {
	log.Printf("[DEBUG] Getting resource ID for type: %s", resourceType)
	id, err := h.resourceRepo.GetResourceIDByType(ctx, resourceType)
	if err != nil {
		log.Printf("[ERROR] Failed to get resource ID for type %s: %v", resourceType, err)
		return "", err
	}
	log.Printf("[DEBUG] Found resource ID %s for type %s", id, resourceType)
	return id, nil
}

func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from header
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		log.Printf("[ERROR] Missing X-User-ID header in request: %s %s", r.Method, r.URL.Path)
		http.Error(w, "Missing X-User-ID header", http.StatusBadRequest)
		return
	}
	log.Printf("[INFO] Handling request from user %s: %s %s", userID, r.Method, r.URL.Path)

	// Check if this is a non-resource path (e.g., /health)
	path := r.URL.Path
	if path == "/health" {
		log.Printf("[INFO] Non-resource path, forwarding directly: %s", path)
		h.proxy.ServeHTTP(w, r)
		return
	}

	// Extract and validate resource information
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	var (
		resourceType string
		resourceID   string
		err          error
	)

	resourceType = parts[0]
	resourceID, err = h.getResourceID(r.Context(), resourceType)
	if err != nil {
		log.Printf("[ERROR] Failed to get resource ID: %v", err)
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	if len(parts) >= 2 {
		resourceID = parts[1]
		log.Printf("[INFO] Accessing specific resource: type=%s, id=%s", resourceType, resourceID)
	} else {
		log.Printf("[INFO] Accessing resource collection: type=%s", resourceType)
	}

	// Only support GET method for view action
	if r.Method != http.MethodGet {
		log.Printf("[ERROR] Unsupported HTTP method: %s", r.Method)
		http.Error(w, "Only GET method is supported", http.StatusMethodNotAllowed)
		return
	}
	action := "view"

	// First, check access without data
	req := model.EvaluationRequest{
		UserID:       userID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Action:       action,
	}

	// Evaluate initial access
	policyResponse, err := h.checkAccess(r, req)
	if err != nil {
		log.Printf("[ERROR] Failed to check access: %v", err)
		http.Error(w, fmt.Sprintf("Failed to check access: %v", err), http.StatusInternalServerError)
		return
	}

	if !policyResponse.Allow {
		log.Printf("[INFO] Access denied: user=%s, resourceType=%s, resourceID=%s, action=%s",
			userID, resourceType, resourceID, action)
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Create a response interceptor
	interceptor := &responseInterceptor{
		writer: w,
	}

	// Forward the request and get the response
	originalWriter := interceptor.writer
	h.proxy.ServeHTTP(interceptor, r)

	// Handle response data
	if !interceptor.hasContent {
		interceptor.copyHeadersTo(originalWriter)
		if interceptor.statusCode > 0 {
			originalWriter.WriteHeader(interceptor.statusCode)
		}
		originalWriter.Write(interceptor.body)
		return
	}

	var data map[string][]interface{}
	if err := json.Unmarshal(interceptor.body, &data); err != nil {
		log.Printf("[ERROR] Failed to unmarshal response: %v", err)
		http.Error(originalWriter, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("[DEBUG] Received data from backend: %+v", data)

	// Re-evaluate with data for field-level filtering
	req.Data = data

	policyResponse, err = h.checkAccess(r, req)
	if err != nil {
		log.Printf("[ERROR] Failed to filter data: %v", err)
		http.Error(originalWriter, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Copy headers and write response
	interceptor.copyHeadersTo(originalWriter)
	originalWriter.Header().Set("Content-Type", "application/json")

	if policyResponse.FilteredData == nil {
		log.Printf("[ERROR] No filtered data in policy response")
		http.Error(originalWriter, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(originalWriter).Encode(policyResponse.FilteredData); err != nil {
		log.Printf("[ERROR] Failed to encode filtered data: %v", err)
		http.Error(originalWriter, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("[DEBUG] Successfully filtered and sent response for resource: %s", resourceType)
}

func main() {
	log.Printf("Starting PEP proxy server on port 80")

	// Initialize database connection
	log.Printf("[DEBUG] Connecting to database...")
	conn, err := pgx.Connect(context.Background(), "postgres://postgres:postgres@prp-db:5432/prp?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(context.Background())

	// Test database connection
	var testResult string
	err = conn.QueryRow(context.Background(), "SELECT 'connection_test'").Scan(&testResult)
	if err != nil {
		log.Fatalf("Failed to query database: %v", err)
	}
	log.Printf("[DEBUG] Database connection test successful: %s", testResult)

	// Initialize repository
	log.Printf("[DEBUG] Initializing repository...")
	repo := repository.NewRepository(conn)

	// Initialize proxy handler
	proxyHandler := NewProxyHandler("http://pdp:8081", repo)

	mux := http.NewServeMux()
	mux.Handle("/", proxyHandler)

	server := &http.Server{
		Handler:      mux,
		Addr:         "0.0.0.0:80",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}
