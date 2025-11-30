package posts

import (
	"comments-system/internal/domain"
	"context"
)

type postsUsecase interface {
	CreatePost(ctx context.Context, post domain.Post) (domain.Post, error)
	GetPosts(ctx context.Context, page, pageSize int, searchQuery, sortBy, sortOrder string) ([]domain.Post, int, error)
	GetPostByID(ctx context.Context, id int) (domain.Post, error)
	DeletePost(ctx context.Context, id int) error
}
