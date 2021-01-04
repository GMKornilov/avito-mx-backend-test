package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type salesHandler struct {
	DB sql.DB
}

type offer struct {
	offerId  int
	sellerId int
	name     string
	price    int
	quantity int
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
			newFilter = `seller_id = ?`
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
			newFilter = `offer_id = ?`
			filters = append(filters, newFilter)
			filterVals = append(filterVals, offerId)
		}
	}

	nameQuery := r.URL.Query().Get("query")
	if nameQuery != "" {
		newFilter = `name LIKE '%' || ? || '%'`
		filters = append(filters, newFilter)
		filterVals = append(filterVals, nameQuery)
	}

	query := "SELECT offer_id, seller_id, name, price, quantity FROM offers"
	if len(filters) > 0 {
		query += "WHERE"
		query += strings.Join(filters, " AND ")
	}

	rows, err := h.DB.Query(query, filterVals...)
	if err != nil {
		// TODO: log error
		http.Error(w, "Error processing query in SQL", 500)
		return
	}

	offerRows := make([]*offer, 0)

	for rows.Next() {
		offerRow := &offer{}
		err = rows.Scan(&offerRow.offerId, &offerRow.sellerId, &offerRow.name, &offerRow.price, &offerRow.quantity)
		if err != nil {
			// TODO: log error
			http.Error(w, "Error processing query in SQL", 500)
			return
		}
		offerRows = append(offerRows, offerRow)
	}
	rows.Close()

	responseStr , _ := json.Marshal(offerRows)
	w.Write([]byte(responseStr))
}
