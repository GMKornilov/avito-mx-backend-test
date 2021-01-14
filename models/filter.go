package models

type Filter struct {
	SellerId *int    `json:"seller_id"`
	OfferId  *int    `json:"offer_id"`
	Query    *string `json:"query"`
}

