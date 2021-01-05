package models

import (
	"github.com/tealeg/xlsx/v3"
)

type UploadQueryRow struct {
	Sale      Sale
	Available bool
}

func FromExcelRow(row *xlsx.Row, sellerId int) (*UploadQueryRow, error) {
	offerId, err := row.GetCell(0).Int()
	if err != nil {
		return nil, err
	}

	name := row.GetCell(1).String()

	price, err := row.GetCell(2).Int()
	if err != nil {
		return nil, err
	}

	quantity, err := row.GetCell(3).Int()
	if err != nil {
		return nil, err
	}

	available := row.GetCell(4).Bool()

	result := &UploadQueryRow{
		Sale: Sale{
			SellerId: sellerId,
			OfferId:  offerId,
			Name:     name,
			Price:    price,
			Quantity: quantity,
		},
		Available: available,
	}

	return result, nil
}

func FromExcelSheet(sheet *xlsx.Sheet, sellerId int) ([]*UploadQueryRow, error) {
	var rows []*UploadQueryRow
	err := sheet.ForEachRow(func(row *xlsx.Row) error {
		uploadQueryRow, err := FromExcelRow(row, sellerId)
		if err != nil {
			return err
		}
		rows = append(rows, uploadQueryRow)
		return nil
	})

	if err != nil {
		return nil, err
	}
	return rows, nil
}

func FromExcelFile(file *xlsx.File, sellerId int) ([]*UploadQueryRow, error) {
	var queryRows = []*UploadQueryRow{}
	for _, sheet := range file.Sheet {
		sheetQueryRows, err := FromExcelSheet(sheet, sellerId)
		if err != nil {
			return nil, err
		}
		queryRows = append(queryRows, sheetQueryRows...)
	}
	return queryRows, nil
}
