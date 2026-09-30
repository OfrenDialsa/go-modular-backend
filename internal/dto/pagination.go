package dto

type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type Filter struct {
	Page  int
	Limit int
	Sort
}

type Sort struct {
	SortBy    string
	SortOrder string
}
