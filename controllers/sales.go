package controllers

import (
	"database/sql"
	"encoding/json"
	"github.com/fertilewaif/avito-mx-backend-test/models"
	"github.com/fertilewaif/avito-mx-backend-test/utils"
	"github.com/tealeg/xlsx/v3"
	"io"
	"net/http"
	"os"
	"strconv"
)

type SalesController interface {
	GetSales(w http.ResponseWriter, r *http.Request)
	Upload(w http.ResponseWriter, r *http.Request)
	Close()
}

type salesController struct {
	Sales *models.Sales
}

type uploadRequest struct {
	SellerId int `json:"seller_id"`
	ExcelUrl string `json:"path"`
}

func NewSalesController(DB *sql.DB) SalesController {
	return &salesController{
		Sales: &models.Sales{DB: DB},
	}
}

func (s *salesController) GetSales(w http.ResponseWriter, r *http.Request) {
	filter := models.Filter{}

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
		return
	}

	respJson, _ := json.Marshal(sales)
	w.Write(respJson)
}

func (s *salesController) Upload(w http.ResponseWriter, r *http.Request) {
	var req uploadRequest

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		respError := models.Error{
			Code:    http.StatusBadRequest,
			Message: "Error parsing request body",
		}
		respJson, _ := json.Marshal(respError)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(respJson)
		return
	}

	uploadStatus := models.UploadStatus{}

	download, err := http.Get(req.ExcelUrl)
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
			newUploadQuery, err := models.FromExcelRow(row, req.SellerId)
			if err != nil {
				uploadStatus.QueryErrors++
				return nil
			}
			uploadQuery = *newUploadQuery
			s.ProcessQuery(uploadQuery, &uploadStatus)
			return nil
		})
	}

	respJson, _ := json.Marshal(uploadStatus)
	w.Write(respJson)
}

func (s *salesController) Close() {
	s.Sales.Close()
}

func (s *salesController) ProcessQuery(q models.UploadQueryRow, u *models.UploadStatus) {
	if q.Available {
		// offer is available, we need to insert/update sale data
		sale, err := s.Sales.FindByIdPair(q.Sale.SellerId, q.Sale.OfferId)
		if err != nil && err != sql.ErrNoRows {
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
