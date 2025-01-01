package main

import (
	"fmt"
	"log"
	"poc-opa-access-control-system/internal/pkg"

	_ "github.com/jackc/pgx/v5"
)

func main() {
	// NOTE: HTTP client examples
	httpClient := pkg.NewClient("http://localhost:8080")
	response, err := httpClient.Get("endpoint")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println("Response Status:", response.Status)

	body := map[string]string{"key": "value"}
	response, err = httpClient.Post("endpoint", body)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println("Response Status:", response.Status)

	// NOTE: DB conn examples
	settings := map[string]pkg.DBConfig{
		"foo": {
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			DBName:   "foo",
			SSLMode:  "disable",
		},
		"prp": {
			Host:     "localhost",
			Port:     5433,
			User:     "postgres",
			Password: "postgres",
			DBName:   "prp",
			SSLMode:  "disable",
		},
	}

	manager := pkg.NewDBManager(settings)
	defer manager.CloseAll()

	clientFoo, err := manager.GetClient("foo")
	if err != nil {
		log.Fatalf("Failed to get client for foo: %v", err)
	}

	var result string
	err = clientFoo.QueryRow("SELECT 'Hello, Foo!'").Scan(&result)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	fmt.Println(result)

	clientPrp, err := manager.GetClient("prp")
	if err != nil {
		log.Fatalf("Failed to get client for prp: %v", err)
	}

	err = clientPrp.QueryRow("SELECT 'Hello, PRP!'").Scan(&result)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	fmt.Println(result)
}
