package models

type PackSize struct {
	ID   int64 `json:"id"`
	Size int   `json:"size"`
}

type ListPackSizesResponse struct {
	Packs []PackSize `json:"packs"`
}

type CreatePackSizeRequest struct {
	Size int `json:"size"`
}

type UpdatePackSizeRequest struct {
	Size int `json:"size"`
}
