package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"net/http"
)

func main() {
	connStr := "user=postgres password=postgres dbname=postgres sslmode=disable"

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	handler := &salesHandler{db}
	http.HandleFunc("/offers/", handler.getOffers)
	http.ListenAndServe(":8080", nil)
}
