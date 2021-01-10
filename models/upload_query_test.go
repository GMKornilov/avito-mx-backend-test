package models_test

import (
	"github.com/fertilewaif/avito-mx-backend-test/models"
	"github.com/tealeg/xlsx/v3"
	"reflect"
	"testing"
)

func createRow(rowVals []interface{}) *xlsx.Row {
	sheet, _ := xlsx.NewSheet("test")
	row := sheet.AddRow()

	for _, cellVal := range rowVals {
		newCell := row.AddCell()
		// library bug: when SetValue is executed with bool under cellVal, SetBool isn't executed
		if boolVal, ok := cellVal.(bool); ok {
			newCell.SetBool(boolVal)
		} else {
			newCell.SetValue(cellVal)
		}
	}
	return row
}

func TestFromExcelRow(t *testing.T) {
	rowVals := []interface{}{1, "offer_name", 100, 2, true}
	sellerId := 1

	expectedQuery := &models.UploadQueryRow{
		Sale: models.Sale{
			OfferId:  1,
			SellerId: sellerId,
			Name:     "offer_name",
			Price:    100,
			Quantity: 2,
		},
		Available: true,
	}

	row := createRow(rowVals)

	uploadQuery, err := models.FromExcelRow(row, sellerId)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	if !reflect.DeepEqual(uploadQuery, expectedQuery) {
		t.Errorf("Unexpected value.\nExpected %+v.\nGot %+v", expectedQuery, uploadQuery)
	}
}


func TestBadData(t *testing.T) {
	sellerId := 1
	rowsVals := [][]interface{}{
		{"bad offer_id", "offer_1", 300, 10, true},
		{4, "offer_2", "bad price", 2, false},
		{3, "offer_3", 10, "bad quantity", false},
	}

	for _, rowVal := range rowsVals {
		row := createRow(rowVal)
		_, err := models.FromExcelRow(row, sellerId)
		if err == nil {
			t.Errorf("Error expected. Input data: %+v", rowVal)
		}
	}
}