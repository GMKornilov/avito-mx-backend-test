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
