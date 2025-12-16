package models

type CalculateRequest struct {
	Quantity int `json:"quantity"`
}

type PackAllocation struct {
	Size  int `json:"size"`
	Count int `json:"count"`
}

type CalculateResponse struct {
	Packs []PackAllocation `json:"packs"`
}
