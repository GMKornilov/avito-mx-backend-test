package main

import (
	"net/http"
)

func main() {
	http.ListenAndServe(":8181", http.FileServer(http.Dir("../uploads")))
}
