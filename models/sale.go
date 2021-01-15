package models

import (
	"database/sql"
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
)

type Sale struct {
	OfferId  int    `json:"offer_id"`
	SellerId int    `json:"seller_id"`
	Name     string `json:"name"`
	Price    int    `json:"price"`
	Quantity int    `json:"quantity"`
}

type Sales struct {
	DB *sql.DB
}

func (h *Sales) AddSale(newSale Sale) (int64, error) {
	query := `INSERT INTO sales (seller_id, offer_id, price, name, quantity) VALUES ($1, $2, $3, $4, $5);`
	res, err := h.DB.Exec(query, newSale.SellerId, newSale.OfferId, newSale.Price, newSale.Name, newSale.Quantity)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"sale":  newSale,
			"query": query,
		}).Errorln("Error processing query")

		return 0, err
	}
	rowsInserted, err := res.RowsAffected()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"sale":  newSale,
			"query": query,
		})

		return 0, err
	}
	return rowsInserted, nil
}

func (h *Sales) FindByIdPair(sellerId int, offerId int) (*Sale, error) {
	sale := new(Sale)
	query := `SELECT offer_id, seller_id, name, price, quantity FROM sales WHERE seller_id = $1 AND offer_id = $2`
	err := h.DB.QueryRow(query, sellerId, offerId).Scan(&sale.OfferId, &sale.SellerId, &sale.Name, &sale.Price, &sale.Quantity)
	if err != nil {
		if err == sql.ErrNoRows {
			log.WithFields(log.Fields{
				"query":     query,
				"seller_id": sellerId,
				"offer_id":  offerId,
			}).Warningln("No rows were selected")

			return nil, nil
		}

		log.WithFields(log.Fields{
			"query":     query,
			"error":     err,
			"seller_id": sellerId,
			"offer_id":  offerId,
		})
		return nil, err
	}
	return sale, nil
}

func (h *Sales) UpdateSale(sale Sale) (int64, error) {
	query := `UPDATE sales SET price=$3, name=$4, quantity=$5 WHERE seller_id = $1 AND offer_id = $2;`
	res, err := h.DB.Exec(query, sale.SellerId, sale.OfferId, sale.Price, sale.Name, sale.Quantity)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"query": query,
			"sale":  sale,
		}).Errorln("Error updating sale")

		return 0, err
	}
	rowsUpdated, err := res.RowsAffected()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"query": query,
			"sale":  sale,
		}).Errorln("Error getting amount of affected rows after updating sale")

		return 0, err
	}
	return rowsUpdated, nil
}

func (h *Sales) DeleteByIdPair(sellerId int, offerId int) (int64, error) {
	query := `DELETE FROM sales WHERE seller_id = $1 AND offer_id = $2;`
	res, err := h.DB.Exec(query, sellerId, offerId)
	if err != nil {
		log.WithFields(log.Fields{
			"error":     err,
			"query":     query,
			"seller_id": sellerId,
			"offer_id":  offerId,
		}).Errorln("Error deleting sale from database")

		return 0, err
	}
	rowsDeleted, err := res.RowsAffected()
	if err != nil {
		log.WithFields(log.Fields{
			"error":     err,
			"query":     query,
			"seller_id": sellerId,
			"offer_id":  offerId,
		}).Errorln("Error getting amount of deleted rows")

		return 0, err
	}
	return rowsDeleted, nil
}

func (h *Sales) FindByFilter(filter Filter) ([]Sale, error) {
	var sales []Sale

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
		log.WithFields(log.Fields{
			"error":  err,
			"query":  query,
			"filter": filter,
		}).Errorln("Error selecting with filter")

		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		saleRow := Sale{}

		err := rows.Scan(&saleRow.OfferId, &saleRow.SellerId, &saleRow.Name, &saleRow.Price, &saleRow.Quantity)

		if err != nil {
			if err == sql.ErrNoRows {
				log.WithFields(log.Fields{
					"query":  query,
					"filter": filter,
				}).Warningln("No rows were selected")
				return nil, nil
			}

			log.WithFields(log.Fields{
				"error":  err,
				"query":  query,
				"filter": filter,
			}).Errorln("Error selecting with filter")
			return nil, err
		}

		sales = append(sales, saleRow)
	}

	return sales, nil
}

func (h *Sales) Close() {
	h.DB.Close()
}
