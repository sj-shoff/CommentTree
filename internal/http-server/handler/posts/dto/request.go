package dto

import "github.com/go-playground/validator/v10"

type CreatePostRequest struct {
	Title   string `json:"title" validate:"required,min=1,max=200"`
	Content string `json:"content" validate:"required,min=1,max=10000"`
	Author  string `json:"author" validate:"required,min=2,max=50"`
}

type GetPostsRequest struct {
	Page      int    `query:"page"`
	PageSize  int    `query:"page_size"`
	Search    string `query:"search"`
	SortBy    string `query:"sort_by"`
	SortOrder string `query:"sort_order"`
}

func (r *GetPostsRequest) Validate() error {
	if r.Page < 1 {
		r.Page = 1
	}
	if r.PageSize < 1 {
		r.PageSize = 10
	}
	if r.PageSize > 100 {
		r.PageSize = 100
	}

	validSortFields := map[string]bool{
		"created_at": true,
		"title":      true,
		"id":         true,
	}
	validSortOrders := map[string]bool{
		"asc":  true,
		"desc": true,
	}

	if !validSortFields[r.SortBy] {
		r.SortBy = "created_at"
	}
	if !validSortOrders[r.SortOrder] {
		r.SortOrder = "desc"
	}

	validate := validator.New()
	return validate.Struct(r)
}
