package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/fertilewaif/avito-mx-backend-test/models"
	"github.com/fertilewaif/avito-mx-backend-test/utils"
	"github.com/tealeg/xlsx/v3"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Sales struct {
	DB        *sql.DB
	totalJobs int
}

func NewSalesController(DB *sql.DB) *Sales {
	return &Sales{
		DB:        DB,
		totalJobs: 0,
	}
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
				Code:    http.StatusBadRequest,
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
				Code:    http.StatusBadRequest,
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
		newFilter := fmt.Sprintf(`LOWER(name) LIKE '%%' || LOWER($%d) || '%%'`, len(filters)+1)
		filters = append(filters, newFilter)
		filterVals = append(filterVals, nameQuery)
	}

	query := "SELECT offer_id, seller_id, name, price, quantity FROM sales"
	if len(filters) > 0 {
		query += " WHERE "
		query += strings.Join(filters, " AND ")
	}
	query += ";"

	rows, err := h.DB.Query(query, filterVals...)

	if err != nil {
		// TODO: log error
		respError := models.Error{
			Code:    http.StatusInternalServerError,
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
				Code:    http.StatusInternalServerError,
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

func (h *Sales) Upload(w http.ResponseWriter, r *http.Request) {
	uploadStatus := models.UploadStatus{}

	excelUrl := r.URL.Query().Get("path")

	sellerIdStr := r.URL.Query().Get("seller_id")
	sellerId, err := strconv.Atoi(sellerIdStr)
	if err != nil {
		// TODO: log error
		respError := models.Error{
			Code:    http.StatusBadRequest,
			Message: "Invalid value of seller_id, should be integer",
		}
		respJson, _ := json.Marshal(respError)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(respJson)
		return
	}

	download, err := http.Get(excelUrl)
	if err != nil {
		// TODO: log error
		respError := models.Error{
			Code:    http.StatusBadRequest,
			Message: "Couldn't download xlsx file on given link",
		}
		respJson, _ := json.Marshal(respError)
		w.WriteHeader(respError.Code)
		w.Write(respJson)
		return
	}
	defer download.Body.Close()
	tmpFilePath := "./uploads/" + utils.RandStringRunes(40) + ".xlsx"
	tmpFile, err := os.Create(tmpFilePath)

	if err != nil {
		// TODO: log error
		respError := models.Error{
			Code:    http.StatusInternalServerError,
			Message: "Error creating temporary file",
		}
		respJson, _ := json.Marshal(respError)
		w.WriteHeader(respError.Code)
		w.Write(respJson)
		return
	}

	_, err = io.Copy(tmpFile, download.Body)

	if err != nil {
		// TODO: log error
		respError := models.Error{
			Code:    http.StatusInternalServerError,
			Message: "Error downloading xlsx file",
		}
		respJson, _ := json.Marshal(respError)
		w.WriteHeader(respError.Code)
		w.Write(respJson)
		tmpFile.Close()
		return
	}
	tmpFile.Close()

	wb, err := xlsx.OpenFile(tmpFilePath)

	if err != nil {
		// TODO: log error
		respError := models.Error{
			Code:    http.StatusInternalServerError,
			Message: "Error opening xlsx file(maybe file has wrong format)",
		}
		respJson, _ := json.Marshal(respError)
		w.WriteHeader(respError.Code)
		w.Write(respJson)
		return
	}

	for _, sheet := range wb.Sheets {
		var uploadQuery models.UploadQueryRow
		sheet.ForEachRow(func(row *xlsx.Row) error {
			newUploadQuery, err := models.FromExcelRow(row, sellerId)
			if err != nil {
				uploadStatus.QueryErrors++
				return nil
			}
			uploadQuery = *newUploadQuery
			return nil
		})
		h.processQuery(w, r, uploadQuery, &uploadStatus)
	}

	respJson, _ := json.Marshal(uploadStatus)
	w.Write(respJson)
}

func (h *Sales) processQuery(w http.ResponseWriter, r *http.Request, q models.UploadQueryRow, u *models.UploadStatus) {
	if q.Available {
		// offer is available, we need to insert/update sale data
		available, err := h.checkAvailability(q.Sale.SellerId, q.Sale.OfferId)
		if err != nil {
			// error happened while checking availability, count it as error during processing query
			u.InternalErrors++
			return
		}
		var query string
		if available {
			// there is such sale in db, need to update it
			query = `UPDATE sales SET price=$3, name=$4, quantity=$5 WHERE seller_id = $1 AND offer_id = $2;`
		} else {
			// there is no such sale in db, creating new one
			query = `INSERT INTO sales (seller_id, offer_id, price, name, quantity) VALUES ($1, $2, $3, $4, $5);`
		}
		res, err := h.DB.Exec(query, q.Sale.SellerId, q.Sale.OfferId, q.Sale.Price, q.Sale.Name, q.Sale.Quantity)
		if err != nil {
			// error happened while updating sales
			u.InternalErrors++
			return
		}
		rowsUpdated, err := res.RowsAffected()
		if err != nil {
			u.InternalErrors++
			return
		}
		u.UpdatedSales += rowsUpdated
	} else {
		// offer is unavailable, we need to delete it from db
		query := `DELETE FROM sales WHERE seller_id = $1 AND offer_id = $2;`
		res, err := h.DB.Exec(query, q.Sale.SellerId, q.Sale.OfferId)
		if err != nil {
			u.InternalErrors++
			return
		}
		rowsDeleted, err := res.RowsAffected()
		if err != nil {
			u.InternalErrors++
			return
		}
		u.DeletedSales += rowsDeleted
	}
}

// checks if there is a sale with pair of (sellerId, offerId)
func (h *Sales) checkAvailability(sellerId int, offerId int) (bool, error) {
	var tmpSaleId int
	query := "SELECT sale_id FROM sales WHERE seller_id = $1 AND offer_id = $2;"
	err := h.DB.QueryRow(query, sellerId, offerId).Scan(&tmpSaleId)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		// TODO: log error
		return false, err
	}
	return true, nil
}
