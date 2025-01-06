package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"poc-opa-access-control-system/internal/model"
	"poc-opa-access-control-system/internal/pkg"
)

// PEPServer is a struct that represents a PEP server.
type PEPServer struct {
	proxy *httputil.ReverseProxy
}

var backends = map[string]string{
	"foo.local": "http://foo.local:8080",
}

func getTargetURL(r *http.Request) (*url.URL, error) {
	backend, exists := backends[r.Host]
	if !exists {
		return nil, fmt.Errorf("Backend not found")
	}
	targetURL, err := url.Parse(backend)
	if err != nil {
		return nil, fmt.Errorf("Error parsing URL: %v", err)
	}
	return targetURL, nil
}

// NewPEPServer creates a new PEP server.
func NewPEPServer() *PEPServer {
	director := func(req *http.Request) {
		targetURL, err := getTargetURL(req)
		if err != nil {
			log.Printf("Error getting target URL: %v", err)
			return
		}
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.Header.Set("X-Forwarded-Host", req.Host)
		req.Header.Set("X-Forwarded-Proto", req.URL.Scheme)
	}
	return &PEPServer{
		proxy: &httputil.ReverseProxy{Director: director},
	}
}

// PIP communication communicates with the PIP.
// Get the additional information required for PDP processing.
func pipCommunication(userID string) (statusCode int, userName string, err error) {
	httpClient := pkg.NewClient("http://pip.local:8082")
	body := map[string]string{"id": userID}
	res, err := httpClient.Post("userinfo", body)
	if err != nil {
		return http.StatusInternalServerError, "", fmt.Errorf("Error contacting PIP: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return res.StatusCode, "", fmt.Errorf("Access denied by pip")
	}
	var u model.User
	if err := json.NewDecoder(res.Body).Decode(&u); err != nil {
		return http.StatusInternalServerError, "", fmt.Errorf("Error decoding PIP response: %v", err)
	}
	// NOTE: Get user name from PIP as an additional information required to reference the policy.
	// However, it is only obtained in a pseudo manner and is not used in PDP processing.
	return http.StatusOK, u.Name, nil
}

// PDP Policy communication communicates with the PDP.
// Get the policy information required for PDP processing.
func pdpPolicyCommunication() (statusCode int, policy string, err error) {
	httpClient := pkg.NewClient("http://pdp.local:8081")
	// TODO: Use POST method.
	res, err := httpClient.Get("policy")
	if err != nil {
		return http.StatusInternalServerError, "", fmt.Errorf("Error contacting PDP: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return res.StatusCode, "", fmt.Errorf("Access denied by pdp")
	}

	// Use struct to store the response.
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return http.StatusInternalServerError, "", fmt.Errorf("Error reading PDP response: %v", err)
	}
	return http.StatusOK, string(body), nil
}

// externalServiceCommunication communicates with the external service.
// Get the response from the external service.
func externalServiceCommunication(targetURL string, targetPath string) (statusCode int, rslt string, err error) {
	httpClient := pkg.NewClient(targetURL)
	res, err := httpClient.Get(targetPath)
	if err != nil {
		return http.StatusInternalServerError, "", fmt.Errorf("Error contacting external service: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return res.StatusCode, "", fmt.Errorf("Access denied by external service")
	}

	// TODO: Use struct to store the response.
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return http.StatusInternalServerError, "", fmt.Errorf("Error reading external service response: %v", err)
	}
	return http.StatusOK, string(body), nil
}

// PDP Evaluate communication communicates with the PDP.
// Get the evaluation result from the PDP.
func pdpEvaluateCommunication() (statusCode int, rslt string, err error) {
	httpClient := pkg.NewClient("http://pdp.local:8081")
	res, err := httpClient.Get("evaluation") // TODO: Use POST method if necessary
	if err != nil {
		return http.StatusInternalServerError, "", fmt.Errorf("Error contacting PDP: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return res.StatusCode, "", fmt.Errorf("Access denied by pdp")
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return http.StatusInternalServerError, "", fmt.Errorf("Error reading PDP response: %v", err)
	}
	return http.StatusOK, string(body), nil
}

// ServeHTTP intercepts the request and processes it.
func (s *PEPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Intercepted request to %s %s", r.Method, r.URL.String())

	// NOTE: For simplicity, I use UserID, but if you implement token-based authentication, you can use JWT etc.
	// I think it depends on the implementation what you want to use as the key to reference the policy information.
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		log.Printf("Missing X-User-ID header")
		http.Error(w, "Missing X-User-ID header", http.StatusBadRequest)
		return
	}

	sc, un, err := pipCommunication(userID)
	if err != nil {
		log.Printf("Error in PIP communication: %v", err)
		http.Error(w, "Error in PIP communication", sc)
		return
	}
	fmt.Printf("PIP response: %s\n", un)

	sc, po, err := pdpPolicyCommunication()
	if err != nil {
		log.Printf("Error in PDP communication: %v", err)
		http.Error(w, "Error in PDP communication", sc)
		return
	}
	fmt.Printf("PDP response: %s\n", po)

	targetURL, err := getTargetURL(r)
	if err != nil {
		log.Printf("Backend not found: %v", err)
		http.Error(w, "Backend not found", http.StatusNotFound)
		return
	}
	targetPath := r.URL.Path
	sc, rs, err := externalServiceCommunication(targetURL.String(), targetPath)
	if err != nil {
		log.Printf("Error in external service communication: %v", err)
		http.Error(w, "Error in external service communication", sc)
		return
	}
	fmt.Printf("External service response: %s\n", rs)

	sc, ev, err := pdpEvaluateCommunication()
	if err != nil {
		log.Printf("Error in PDP evaluation communication: %v", err)
		http.Error(w, "Error in PDP evaluation communication", sc)
		return
	}
	fmt.Printf("PDP evaluation response: %s\n", ev)

	// Proxy the request to the external service and return the response directly to the client
	s.proxy.ServeHTTP(w, r)
}

func main() {
	pepServer := NewPEPServer()

	log.Println("PEP server is running on :80")

	if err := http.ListenAndServe(":80", pepServer); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
