package dto

import (
	"comments-system/internal/domain"
	"time"
)

type PostResponse struct {
	ID            int       `json:"id"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	Author        string    `json:"author"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	CommentsCount int       `json:"comments_count"`
}

type PostsResponse struct {
	Posts    []PostResponse `json:"posts"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	HasNext  bool           `json:"has_next"`
	HasPrev  bool           `json:"has_prev"`
}

func FromDomainPost(post domain.Post) PostResponse {
	return PostResponse{
		ID:            post.ID,
		Title:         post.Title,
		Content:       post.Content,
		Author:        post.Author,
		CreatedAt:     post.CreatedAt,
		UpdatedAt:     post.UpdatedAt,
		CommentsCount: post.CommentsCount,
	}
}

func FromDomainPosts(posts []domain.Post) []PostResponse {
	responses := make([]PostResponse, len(posts))
	for i, post := range posts {
		responses[i] = FromDomainPost(post)
	}
	return responses
}
