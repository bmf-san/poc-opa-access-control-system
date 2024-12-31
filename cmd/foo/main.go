package main

import (
	"net/http"
)

func handleUsers(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement Later - fetch users from the database
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
	w.Write([]byte("Hello, Users!"))
}

func main() {
	http.HandleFunc("/users", handleUsers)
	http.ListenAndServe(":8080", nil)
}
