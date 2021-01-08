package main

import (
	"database/sql"
	"fmt"
	"github.com/fertilewaif/avito-mx-backend-test/controllers"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
)

const (
	PORT = 5432
)

func initDB(username, password, database string) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"postgresql://%s:%s@database:%d/%s?sslmode=disable",
		username, password, PORT, database)

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	err = conn.Ping()
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func main() {
	dbUser, dbPassword, dbName :=
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB")

	db, err := initDB(dbUser, dbPassword, dbName)

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
