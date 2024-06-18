package models

type ProductIDAndBody struct {
	ProductID
	ProductBody
}

type ProductID struct {
	ID int `json:"id"`
}

type ProductBody struct {
	Name   string `json:"name,required"`
	Price  int    `json:"price"`
	Amount int    `json:"amount"`
}
