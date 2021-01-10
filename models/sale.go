package models

type Sale struct {
	OfferId  int    `json:"offer_id"`
	SellerId int    `json:"seller_id"`
	Name     string `json:"name"`
	Price    int    `json:"price"`
	Quantity int    `json:"quantity"`
}
