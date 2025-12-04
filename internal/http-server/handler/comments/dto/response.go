package dto

import (
	"comments-system/internal/domain"
	"time"
)

type CommentResponse struct {
	ID        int               `json:"id"`
	ParentID  *int              `json:"parent_id,omitempty"`
	Content   string            `json:"content"`
	Author    string            `json:"author"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Children  []CommentResponse `json:"children,omitempty"`
}

type CommentsResponse struct {
	Comments []CommentResponse `json:"comments"`
	Total    int               `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	HasNext  bool              `json:"has_next"`
	HasPrev  bool              `json:"has_prev"`
}

func FromDomainComment(comment domain.Comment) CommentResponse {
	resp := CommentResponse{
		ID:        comment.ID,
		ParentID:  comment.ParentID,
		Content:   comment.Content,
		Author:    comment.Author,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}

	if len(comment.Children) > 0 {
		resp.Children = make([]CommentResponse, len(comment.Children))
		for i, child := range comment.Children {
			resp.Children[i] = FromDomainComment(child)
		}
	}

	return resp
}

func FromDomainComments(comments []domain.Comment) []CommentResponse {
	responses := make([]CommentResponse, len(comments))
	for i, comment := range comments {
		responses[i] = FromDomainComment(comment)
	}
	return responses
}

func FromDomainCommentTree(tree domain.CommentTree) CommentsResponse {
	return CommentsResponse{
		Comments: FromDomainComments(tree.Comments),
		Total:    tree.Total,
		Page:     tree.Page,
		PageSize: tree.PageSize,
		HasNext:  tree.HasNext,
		HasPrev:  tree.HasPrev,
	}
}
