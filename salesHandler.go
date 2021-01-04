package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type salesHandler struct {
	DB *sql.DB
}

type offer struct {
	OfferId  int    `json:"offer_id"`
	SellerId int    `json:"seller_id"`
	Name     string `json:"name"`
	Price    int    `json:"price"`
	Quantity int    `json:"quantity"`
}

func (h *salesHandler) getOffers(w http.ResponseWriter, r *http.Request) {
	var filters []string
	var filterVals []interface{}
	var newFilter string

	sellerIdStr := r.URL.Query().Get("seller_id")
	if sellerIdStr != "" {
		sellerId, err := strconv.Atoi(sellerIdStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid value of seller_id, must be integer"))
			return
		} else {
			newFilter = fmt.Sprintf("seller_id = $%d", len(filters)+1)
			filters = append(filters, newFilter)
			filterVals = append(filterVals, sellerId)
		}
	}

	offerIdStr := r.URL.Query().Get("offer_id")
	if offerIdStr != "" {
		offerId, err := strconv.Atoi(sellerIdStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid value of offer_id, must be integer"))
			return
		} else {
			newFilter = fmt.Sprintf("offer_id = $%d", len(filters)+1)
			filters = append(filters, newFilter)
			filterVals = append(filterVals, offerId)
		}
	}

	nameQuery := r.URL.Query().Get("query")
	if nameQuery != "" {
		newFilter = fmt.Sprintf(`Name LIKE '%' || $%d || '%'`, len(filters)+1)
		filters = append(filters, newFilter)
		filterVals = append(filterVals, nameQuery)
	}

	query := "SELECT offer_id, seller_id, Name, Price, Quantity FROM offers"
	if len(filters) > 0 {
		query += " WHERE "
		query += strings.Join(filters, " AND ")
	}
	query += ";"

	rows, err := h.DB.Query(query, filterVals...)

	if err != nil {
		// TODO: log error
		http.Error(w, "Error processing query in SQL", 500)
		return
	}

	var offerRows []offer

	for rows.Next() {
		offerRow := offer{}
		err = rows.Scan(&offerRow.OfferId, &offerRow.SellerId, &offerRow.Name, &offerRow.Price, &offerRow.Quantity)
		if err != nil {
			// TODO: log error
			http.Error(w, "Error processing query in SQL rows", 500)
			return
		}
		offerRows = append(offerRows, offerRow)
	}
	rows.Close()

	responseStr, _ := json.Marshal(offerRows)
	w.Write(responseStr)
}
