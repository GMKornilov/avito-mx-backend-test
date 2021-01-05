package models_test

import (
	"github.com/fertilewaif/avito-mx-backend-test/models"
	"github.com/tealeg/xlsx/v3"
	"reflect"
	"testing"
)

func createSheet(rows [][]interface{}) *xlsx.Sheet {
	sheet, _ := xlsx.NewSheet("test")
	for _, rowVal := range rows {
		newRow := sheet.AddRow()

		for _, cellVal := range rowVal {
			newCell := newRow.AddCell()
			// library bug: when SetValue is executed with bool under cellVal, SetBool should be executed
			if boolVal, ok := cellVal.(bool); ok {
				newCell.SetBool(boolVal)
			} else {
				newCell.SetValue(cellVal)
			}
		}
	}
	return sheet
}

func createRow(rowVals []interface{}) *xlsx.Row {
	sheet, _ := xlsx.NewSheet("test")
	row := sheet.AddRow()

	for _, cellVal := range rowVals {
		newCell := row.AddCell()
		// library bug: when SetValue is executed with bool under cellVal, SetBool should be executed
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

func TestFromExcelSheet(t *testing.T) {
	sellerId := 1
	rowsVals := [][]interface{}{
		{1, "offer_1", 300, 10, true},
		{4, "offer_2", 150, 2, false},
		{3, "offer_3", 10, 1, false},
	}

	expectedQueries := []*models.UploadQueryRow{
		{
			Sale: models.Sale{
				OfferId:  1,
				SellerId: sellerId,
				Name:     "offer_1",
				Price:    300,
				Quantity: 10,
			},
			Available: true,
		},
		{
			Sale: models.Sale{
				OfferId:  4,
				SellerId: sellerId,
				Name:     "offer_2",
				Price:    150,
				Quantity: 2,
			},
			Available: false,
		},
		{
			Sale: models.Sale{
				OfferId:  3,
				SellerId: sellerId,
				Name:     "offer_3",
				Price:    10,
				Quantity: 1,
			},
			Available: false,
		},
	}

	sheet := createSheet(rowsVals)

	uploadQueries, err := models.FromExcelSheet(sheet, sellerId)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	if !reflect.DeepEqual(uploadQueries, expectedQueries) {
		t.Errorf("Unexpected value.\nExpected: %+v with length %d\nGot: %+v with length %d",
			expectedQueries, len(expectedQueries), uploadQueries, len(uploadQueries))
	}
}
