package controllers

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"github.com/fertilewaif/avito-mx-backend-test/models"
	"github.com/fertilewaif/avito-mx-backend-test/utils"
	"github.com/tealeg/xlsx/v3"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
)

type Worker interface {
	StartJob(url string, sellerId int) string
	GetJobStatus(jobId string) UploadStatus
	FinishJob(jobId string)
}

type UploadStatus struct {
	Ready        bool                `json:"ready"`
	UploadResult *models.UploadResult `json:"upload_result,omitempty"`
	Error        *models.Error        `json:"error,omitempty"`
}

type worker struct {
	totalJobs int
	sales     *models.Sales
	statuses  map[string]*UploadStatus
	mutex     sync.RWMutex
}

func NewWorker(sales *models.Sales) Worker {
	return &worker{
		totalJobs: 0,
		sales:     sales,
		statuses:  make(map[string]*UploadStatus),
		mutex:     sync.RWMutex{},
	}
}

func (w *worker) generateJobId() string {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.totalJobs++
	byteValue := []byte(strconv.Itoa(w.totalJobs))
	return fmt.Sprintf("%x", md5.Sum(byteValue))
}

func (w *worker) processDownload(url string, sellerId int, uploadStatus *UploadStatus) {
	download, err := http.Get(url)

	if err != nil {
		uploadStatus.Ready = true
		uploadStatus.Error = &models.Error{
			Code:    http.StatusBadRequest,
			Message: "Couldn't download xlsx file on given link",
		}
		return
	}

	tmpFilePath := "./uploads/" + utils.RandStringRunes(40) + ".xlsx"
	tmpFile, err := os.Create(tmpFilePath)

	if err != nil {
		download.Body.Close()

		uploadStatus.Ready = true
		uploadStatus.Error = &models.Error{
			Code:    http.StatusInternalServerError,
			Message: "Error creating temporary file",
		}
		return
	}

	_, err = io.Copy(tmpFile, download.Body)
	download.Body.Close()

	if err != nil {
		uploadStatus.Ready = true
		uploadStatus.Error = &models.Error{
			Code:    http.StatusInternalServerError,
			Message: "Error downloading xlsx file",
		}
		return
	}

	wb, err := xlsx.OpenFile(tmpFilePath)

	if err != nil {
		uploadStatus.Ready = true
		uploadStatus.Error = &models.Error{
			Code:    http.StatusInternalServerError,
			Message: "Error opening xlsx file(maybe file has wrong format)",
		}
		return
	}

	w.processFile(wb, sellerId, uploadStatus)
}

func (w *worker) processFile(excelFile *xlsx.File, sellerId int, uploadStatus *UploadStatus) {
	for _, sheet := range excelFile.Sheets {
		var uploadQuery models.UploadQueryRow
		sheet.ForEachRow(func(row *xlsx.Row) error {
			newUploadQuery, err := models.FromExcelRow(row, sellerId)
			if err != nil {
				uploadStatus.UploadResult.QueryErrors++
				return nil
			}
			uploadQuery = *newUploadQuery
			w.processQuery(uploadQuery, uploadStatus.UploadResult)
			return nil
		})
	}
	uploadStatus.Ready = true
}

func (w *worker) processQuery(q models.UploadQueryRow, u *models.UploadResult) {
	if q.Available {
		// offer is available, we need to insert/update sale data
		sale, err := w.sales.FindByIdPair(q.Sale.SellerId, q.Sale.OfferId)
		if err != nil && err != sql.ErrNoRows {
			// error happened while checking availability, count it as error during processing query
			u.InternalErrors++
			return
		}
		if sale != nil {
			// there is such sale in db, need to update it
			rowsUpdated, err := w.sales.UpdateSale(q.Sale)
			if err != nil {
				u.InternalErrors++
				return
			}
			u.UpdatedSales += rowsUpdated
		} else {
			// there is no such sale in db, creating new one
			rowsCreated, err := w.sales.AddSale(q.Sale)
			if err != nil {
				u.InternalErrors++
				return
			}
			u.CreatedSales += rowsCreated
		}
	} else {
		// offer is unavailable, we need to delete it from db
		rowsDeleted, err := w.sales.DeleteByIdPair(q.Sale.SellerId, q.Sale.OfferId)

		if err != nil {
			u.InternalErrors++
			return
		}

		u.DeletedSales += rowsDeleted
	}
}

func (w *worker) StartJob(url string, sellerId int) string {
	newJobId := w.generateJobId()
	newUploadStatus := &UploadStatus{
		Ready:        false,
		UploadResult: &models.UploadResult{
			CreatedSales:   0,
			UpdatedSales:   0,
			DeletedSales:   0,
			QueryErrors:    0,
			InternalErrors: 0,
		},
		Error:        nil,
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.statuses[newJobId] = newUploadStatus
	go w.processDownload(url, sellerId, newUploadStatus)

	return newJobId
}


func (w *worker) GetJobStatus(jobId string) UploadStatus {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	status, ok := w.statuses[jobId]
	if !ok {
		return UploadStatus{false, nil, nil}
	}
	return *status
}

func (w *worker) FinishJob(jobId string) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if _, ok := w.statuses[jobId]; ok {
		delete(w.statuses, jobId)
	}
}
