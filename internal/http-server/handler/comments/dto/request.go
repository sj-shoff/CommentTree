package dto

import (
	"github.com/go-playground/validator/v10"
)

type CreateCommentRequest struct {
	PostID   int    `json:"post_id" validate:"required,min=1"`
	ParentID *int   `json:"parent_id,omitempty"`
	Content  string `json:"content" validate:"required,min=1,max=1000"`
	Author   string `json:"author" validate:"required,min=2,max=50"`
}

type GetCommentsRequest struct {
	PostID    int    `query:"post_id" validate:"required,min=1"`
	ParentID  *int   `query:"parent"`
	Page      int    `query:"page"`
	PageSize  int    `query:"page_size"`
	Search    string `query:"search"`
	SortBy    string `query:"sort_by"`
	SortOrder string `query:"sort_order"`
}

func (r *GetCommentsRequest) Bind() error {
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
		"updated_at": true,
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

	return nil
}

func (r *GetCommentsRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
