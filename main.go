package main

import (
	"database/sql"
	"github.com/fertilewaif/avito-mx-backend-test/controllers"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

func main() {
	connStr := "user=postgres password=postgres dbname=postgres sslmode=disable"

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	handler := controllers.NewSalesController(db)
	http.HandleFunc("/offers/", handler.GetSales)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
	handler.Close()
}
