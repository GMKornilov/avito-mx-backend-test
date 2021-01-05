package main

import (
	"database/sql"
	"github.com/fertilewaif/avito-mx-backend-test/controllers"
	_ "github.com/lib/pq"
	"net/http"
)

func main() {
	connStr := "user=postgres password=postgres dbname=postgres sslmode=disable"

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	handler := &controllers.SalesHandler{DB: db}
	http.HandleFunc("/offers/", handler.GetSales)
	http.ListenAndServe(":8080", nil)
}
