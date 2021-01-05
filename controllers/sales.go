package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/fertilewaif/avito-mx-backend-test/models"
	"net/http"
	"strconv"
	"strings"
)

type Sales struct {
	DB *sql.DB
}

func (h *Sales) GetSales(w http.ResponseWriter, r *http.Request) {
	var filters []string
	var filterVals []interface{}

	sellerIdStr := r.URL.Query().Get("seller_id")
	if sellerIdStr != "" {
		sellerId, err := strconv.Atoi(sellerIdStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			respError := models.Error{
				Message: "Invalid value of seller_id, must be integer",
				Code: http.StatusBadRequest,
			}
			respJson, _ := json.Marshal(respError)
			w.Write(respJson)
			return
		} else {
			newFilter := fmt.Sprintf("seller_id = $%d", len(filters)+1)
			filters = append(filters, newFilter)
			filterVals = append(filterVals, sellerId)
		}
	}

	offerIdStr := r.URL.Query().Get("offer_id")
	if offerIdStr != "" {
		offerId, err := strconv.Atoi(sellerIdStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			respError := models.Error{
				Message: "Invalid value of offer_id, must be integer",
				Code: http.StatusBadRequest,
			}
			respJson, _ := json.Marshal(respError)
			w.Write(respJson)
			return
		} else {
			newFilter := fmt.Sprintf("offer_id = $%d", len(filters)+1)
			filters = append(filters, newFilter)
			filterVals = append(filterVals, offerId)
		}
	}

	nameQuery := r.URL.Query().Get("query")
	if nameQuery != "" {
		newFilter := fmt.Sprintf(`Name LIKE '%%' || $%d || '%%'`, len(filters)+1)
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
		respError := models.Error{
			Code: http.StatusInternalServerError,
			Message: "Error processing query in SQL",
		}
		respJson, _ := json.Marshal(respError)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(respJson)
		return
	}

	var offerRows []models.Sale

	for rows.Next() {
		offerRow := models.Sale{}
		err = rows.Scan(&offerRow.OfferId, &offerRow.SellerId, &offerRow.Name, &offerRow.Price, &offerRow.Quantity)
		if err != nil {
			// TODO: log error
			respError := models.Error{
				Code: http.StatusInternalServerError,
				Message: "Error parsing SQL query result",
			}
			respJson, _ := json.Marshal(respError)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(respJson)
			return
		}
		offerRows = append(offerRows, offerRow)
	}
	rows.Close()

	responseStr, _ := json.Marshal(offerRows)
	w.Write(responseStr)
}
