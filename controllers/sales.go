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
	DB *sql.DB
}

type filter struct {
	SellerId *int    `json:"seller_id"`
	OfferId  *int    `json:"offer_id"`
	Query    *string `json:"query"`
}

func (h *Sales) AddSale(newSale models.Sale) (int64, error) {
	query := `INSERT INTO sales (seller_id, offer_id, price, name, quantity) VALUES ($1, $2, $3, $4, $5);`
	res, err := h.DB.Exec(query, newSale.SellerId, newSale.OfferId, newSale.Price, newSale.Name, newSale.Quantity)
	if err != nil {
		return 0, nil
	}
	rowsInserted, err := res.RowsAffected()
	if err != nil {
		return 0, nil
	}
	return rowsInserted, nil
}

func (h *Sales) FindByIdPair(sellerId int, offerId int) (*models.Sale, error) {
	sale := new(models.Sale)
	query := `SELECT offer_id, seller_id, name, price, quantity FROM sales WHERE seller_id = $1 AND offer_id = $2`
	err := h.DB.QueryRow(query, sellerId, offerId).Scan(&sale.OfferId, &sale.SellerId, &sale.Name, &sale.Price, &sale.Quantity)
	if err != nil {
		return &models.Sale{}, err
	}
	return sale, nil
}

func (h *Sales) UpdateSale(sale models.Sale) (int64, error) {
	query := `UPDATE sales SET price=$3, name=$4, quantity=$5 WHERE seller_id = $1 AND offer_id = $2;`
	res, err := h.DB.Exec(query, sale.SellerId, sale.OfferId, sale.Price, sale.Name, sale.Quantity)
	if err != nil {
		return 0, err
	}
	rowsUpdated, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rowsUpdated, nil
}

func (h *Sales) DeleteByIdPair(sellerId int, offerId int) (int64, error) {
	query := `DELETE FROM sales WHERE seller_id = $1 AND offer_id = $2;`
	res, err := h.DB.Exec(query, sellerId, offerId)
	if err != nil {
		return 0, err
	}
	rowsDeleted, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rowsDeleted, nil
}

func (h *Sales) FindByFilter(filter filter) ([]models.Sale, error) {
	var sales []models.Sale

	var filters []string
	var filterVals []interface{}

	if filter.SellerId != nil {
		newFilter := fmt.Sprintf("seller_id = $%d", len(filters)+1)
		filters = append(filters, newFilter)
		filterVals = append(filterVals, *filter.SellerId)
	}

	if filter.OfferId != nil {
		newFilter := fmt.Sprintf("offer_id = $%d", len(filters)+1)
		filters = append(filters, newFilter)
		filterVals = append(filterVals, *filter.OfferId)
	}

	if filter.Query != nil {
		newFilter := fmt.Sprintf(`LOWER(name) LIKE '%%' || LOWER($%d) || '%%'`, len(filters)+1)
		filters = append(filters, newFilter)
		filterVals = append(filterVals, *filter.Query)
	}

	query := `SELECT offer_id, seller_id, name, price, quantity FROM sales`
	if len(filters) > 0 {
		query += " WHERE "
		query += strings.Join(filters, " AND ")
	}
	query += ";"

	rows, err := h.DB.Query(query, filterVals...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		saleRow := models.Sale{}

		err := rows.Scan(&saleRow.OfferId, &saleRow.SellerId, &saleRow.Name, &saleRow.Price, &saleRow.Quantity)

		if err != nil {
			return nil, err
		}

		sales = append(sales, saleRow)
	}
	rows.Close()

	return sales, nil
}

type SalesController interface {
	GetSales(w http.ResponseWriter, r *http.Request)
	Upload(w http.ResponseWriter, r *http.Request)
}

type salesController struct {
	Sales *Sales
}

func NewSalesController(DB *sql.DB) SalesController {
	return &salesController{
		Sales: &Sales{DB: DB},
	}
}

func (s *salesController) GetSales(w http.ResponseWriter, r *http.Request) {
	filter := filter{}

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
			filter.SellerId = &sellerId
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
			filter.OfferId = &offerId
		}
	}

	nameQuery := r.URL.Query().Get("query")
	if nameQuery != "" {
		filter.Query = &nameQuery
	}

	sales, err := s.Sales.FindByFilter(filter)

	if err != nil {
		respErr := models.Error{
			Code:    http.StatusInternalServerError,
			Message: "Error processing query",
		}
		respJson, _ := json.Marshal(respErr)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(respJson)
	}

	respJson, _ := json.Marshal(sales)
	w.Write(respJson)
}

func (s *salesController) Upload(w http.ResponseWriter, r *http.Request) {
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
		s.ProcessQuery(uploadQuery, &uploadStatus)
	}

	respJson, _ := json.Marshal(uploadStatus)
	w.Write(respJson)
}

func (s *salesController) ProcessQuery(q models.UploadQueryRow, u *models.UploadStatus) {
	if q.Available {
		// offer is available, we need to insert/update sale data
		sale, err := s.Sales.FindByIdPair(q.Sale.SellerId, q.Sale.OfferId)
		if err != nil {
			// error happened while checking availability, count it as error during processing query
			u.InternalErrors++
			return
		}
		if sale != nil {
			// there is such sale in db, need to update it
			rowsUpdated, err := s.Sales.UpdateSale(q.Sale)
			if err != nil {
				u.InternalErrors++
				return
			}
			u.UpdatedSales += rowsUpdated
		} else {
			// there is no such sale in db, creating new one
			rowsCreated, err := s.Sales.AddSale(q.Sale)
			if err != nil {
				u.InternalErrors++
				return
			}
			u.CreatedSales += rowsCreated
		}
	} else {
		// offer is unavailable, we need to delete it from db
		rowsDeleted, err := s.Sales.DeleteByIdPair(q.Sale.SellerId, q.Sale.OfferId)

		if err != nil {
			u.InternalErrors++
			return
		}
		u.DeletedSales += rowsDeleted
	}
}
