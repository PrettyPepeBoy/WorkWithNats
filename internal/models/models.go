package models

type ProductIDAndBody struct {
	ProductID
	ProductBody
}

type ProductID struct {
	ID int `json:"id"`
}

type ProductBody struct {
	Name     string `json:"name,required"`
	Category string `json:"category,required"`
	Location string `json:"location,omitempty"`
	Color    string `json:"color,omitempty"`
	Price    int    `json:"price,omitempty"`
	Amount   int    `json:"amount,omitempty"`
}
