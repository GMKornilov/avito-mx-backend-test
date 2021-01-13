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
	GetJobStatus(w http.ResponseWriter, r *http.Request)
	Close()
}

type salesController struct {
	Sales  *models.Sales
	Worker Worker
}

type uploadRequest struct {
	SellerId int    `json:"seller_id"`
	ExcelUrl string `json:"path"`
}

func NewSalesController(DB *sql.DB) SalesController {
	sales := &models.Sales{DB: DB}
	return &salesController{
		Sales:  sales,
		Worker: NewWorker(sales),
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

	jobId := s.Worker.StartJob(wb, req.SellerId)

	respJson, _ := json.Marshal(struct {
		JobId string `json:"job_id"`
	}{JobId: jobId})

	w.Write(respJson)
}

func (s *salesController) GetJobStatus(w http.ResponseWriter, r *http.Request) {
	jobId := r.URL.Query().Get("job_id")
	q := s.Worker.GetJobStatus(jobId)
	if !q.Ready {
		q.UploadResult = nil
	} else {
		s.Worker.FinishJob(jobId)
	}
	respJson, _ := json.Marshal(q)
	w.Write(respJson)
}

func (s *salesController) Close() {
	s.Sales.Close()
}
