package models

type UploadStatus struct {
	CreatedSales   int64 `json:"created_sales"`
	UpdatedSales   int64 `json:"updated_sales"`
	DeletedSales   int64 `json:"deleted_sales"`
	QueryErrors    int64 `json:"query_errors"`
	InternalErrors int64 `json:"internal_errors"`
}
