package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5"
	"github.com/open-policy-agent/opa/v1/rego"
)

// GET /policy
func handlePolicy(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement Later - fetch policy from the database by user-id
	// settings := map[string]pkg.DBConfig{
	// 	"foo": {
	// 		Host:     "localhost",
	// 		Port:     5432,
	// 		User:     "postgres",
	// 		Password: "postgres",
	// 		DBName:   "foo",
	// 		SSLMode:  "disable",
	// 	},
	// 	"prp": {
	// 		Host:     "localhost",
	// 		Port:     5433,
	// 		User:     "postgres",
	// 		Password: "postgres",
	// 		DBName:   "prp",
	// 		SSLMode:  "disable",
	// 	},
	// }

	// manager := pkg.NewDBManager(settings)
	// defer manager.CloseAll()

	// clientFoo, err := manager.GetClient("foo")
	// if err != nil {
	// 	log.Fatalf("Failed to get client for foo: %v", err)
	// }

	// var result string
	// err = clientFoo.QueryRow("SELECT 'Hello, Foo!'").Scan(&result)
	// if err != nil {
	// 	log.Fatalf("Query failed: %v", err)
	// }
	// fmt.Println(result)

	// clientPrp, err := manager.GetClient("prp")
	// if err != nil {
	// 	log.Fatalf("Failed to get client for prp: %v", err)
	// }

	// err = clientPrp.QueryRow("SELECT 'Hello, PRP!'").Scan(&result)
	// if err != nil {
	// 	log.Fatalf("Query failed: %v", err)
	// }
	// fmt.Println(result)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Policy"))
}

// POST /evaluation
func handleEvaluation(w http.ResponseWriter, r *http.Request) {
	policy, err := os.ReadFile("./cmd/pdp/policy/example.rego")
	if err != nil {
		log.Fatalf("Error reading policy file: %v", err)
	}

	rg := rego.New(
		rego.Query("data.policy.allow"),
		rego.Input(map[string]interface{}{
			"method": "GET",
			"path":   "/allowed",
		}),
		rego.Module("example.rego", string(policy)),
	)

	rs, err := rg.Eval(context.Background())
	if err != nil {
		fmt.Println("Error evaluating policy:", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Println("Policy evaluation result:", rs)
	w.Write([]byte("Evaluation"))
}

func main() {
	http.HandleFunc("/policy", handlePolicy)
	http.HandleFunc("/evaluation", handleEvaluation)

	http.ListenAndServe(":8081", nil)
}
