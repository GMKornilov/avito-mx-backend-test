package models_test

import (
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/fertilewaif/avito-mx-backend-test/models"
	"reflect"
	"testing"
)

var sale = &models.Sale{
	OfferId:  1,
	SellerId: 10,
	Name:     "Sales test",
	Price:    300,
	Quantity: 100500,
}

func NewMock() (*sql.DB, sqlmock.Sqlmock) {
	db, mock, _ := sqlmock.New()
	return db, mock
}

func TestSales_AddSale(t *testing.T) {
	db, mock := NewMock()
	sales := models.Sales{DB: db}
	defer sales.Close()

	query := `INSERT INTO sales \(seller_id, offer_id, price, name, quantity\) VALUES \(\$1, \$2, \$3, \$4, \$5\);`
	mock.ExpectExec(query).WithArgs(sale.SellerId, sale.OfferId, sale.Price, sale.Name, sale.Quantity).WillReturnResult(sqlmock.NewResult(0, 1))

	rowsInserted, err := sales.AddSale(*sale)

	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	if rowsInserted != 1 {
		t.Errorf("Invalid amount of new rows, expected %d, got %d", 1, rowsInserted)
	}
}

func TestSales_AddSaleError(t *testing.T) {
	db, mock := NewMock()
	sales := models.Sales{DB: db}
	defer sales.Close()

	query := `INSERT INTO sales \(seller_id, offer_id, price, name, quantity\) VALUES \(\$1, \$2, \$3, \$4, \$5\);`
	mock.ExpectExec(query).WithArgs(sale.SellerId, sale.OfferId, sale.Price, sale.Name, sale.Quantity).
		WillReturnError(fmt.Errorf("test error"))

	_, err := sales.AddSale(*sale)

	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestSales_DeleteByIdPair(t *testing.T) {
	db, mock := NewMock()
	sales := models.Sales{DB: db}
	defer sales.Close()

	query := `DELETE FROM sales WHERE seller_id \= \$1 AND offer_id \= \$2;`
	mock.ExpectExec(query).WithArgs(sale.SellerId, sale.OfferId).WillReturnResult(sqlmock.NewResult(0, 1))

	rowsDeleted, err := sales.DeleteByIdPair(sale.SellerId, sale.OfferId)

	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	if rowsDeleted != 1 {
		t.Errorf("Invalid amount of new rows, expected %d, got %d", 1, rowsDeleted)
	}
}

func TestSales_DeleteByIdPairError(t *testing.T) {
	db, mock := NewMock()
	sales := models.Sales{DB: db}
	defer sales.Close()

	query := `DELETE FROM sales WHERE seller_id \= \$1 AND offer_id \= \$2;`
	mock.ExpectExec(query).WithArgs(sale.SellerId, sale.OfferId).WillReturnError(fmt.Errorf("test error"))

	_, err := sales.DeleteByIdPair(sale.SellerId, sale.OfferId)

	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestSales_FindByIdPair(t *testing.T) {
	db, mock := NewMock()
	sales := models.Sales{DB: db}
	defer sales.Close()

	query := `SELECT offer_id, seller_id, name, price, quantity FROM sales WHERE seller_id \= \$1 AND offer_id \= \$2`
	rows := sqlmock.NewRows([]string{"offer_id", "seller_id", "name", "price", "quantity"}).
		AddRow(sale.OfferId, sale.SellerId, sale.Name, sale.Price, sale.Quantity)
	mock.ExpectQuery(query).WithArgs(sale.SellerId, sale.OfferId).WillReturnRows(rows)

	resSale, err := sales.FindByIdPair(sale.SellerId, sale.OfferId)

	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	if !reflect.DeepEqual(resSale, sale) {
		t.Errorf("Invalid result, expected %+v, got %+v", sale, resSale)
	}
}

func TestSales_FindByFilter(t *testing.T) {
	db, mock := NewMock()
	sales := models.Sales{DB: db}
	defer sales.Close()

	offerId := 2
	filterQuery := "test"
	filter := models.Filter{
		SellerId: nil,
		OfferId:  &offerId,
		Query:    &filterQuery,
	}

	query := `SELECT offer_id, seller_id, name, price, quantity FROM sales WHERE offer_id \= \$1 AND LOWER\(name\) LIKE '%' \|\| LOWER\(\$2\) \|\| '%';`
	rows := sqlmock.NewRows([]string{"offer_id", "seller_id", "name", "price", "quantity"}).
		AddRow(sale.OfferId, sale.SellerId, sale.Name, sale.Price, sale.Quantity)
	mock.ExpectQuery(query).WillReturnRows(rows)

	resSale, err := sales.FindByFilter(filter)

	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	saleSlice := []models.Sale{*sale}

	if !reflect.DeepEqual(resSale, saleSlice) {
		t.Errorf("Invalid result, expected %+v, got %+v", saleSlice, resSale)
	}
}

func TestSales_UpdateSale(t *testing.T) {
	db, mock := NewMock()
	sales := models.Sales{DB: db}
	defer sales.Close()

	query := `UPDATE sales SET price\=\$3, name\=\$4, quantity\=\$5 WHERE seller_id \= \$1 AND offer_id \= \$2;`
	mock.ExpectExec(query).WithArgs(sale.SellerId, sale.OfferId, sale.Price, sale.Name, sale.Quantity).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rowsUpdated, err := sales.UpdateSale(*sale)

	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	if rowsUpdated != 1 {
		t.Errorf("Invalid rows updated, expected %d, got %d", 1, rowsUpdated)
	}
}
