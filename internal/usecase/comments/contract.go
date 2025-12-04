package comments_usecase

import (
	"context"

	"comments-system/internal/domain"
)

type commentsRepo interface {
	Create(ctx context.Context, comment domain.Comment) (domain.Comment, error)
	GetByParent(ctx context.Context, parentID *int, page, pageSize int, searchQuery, sortBy, sortOrder string) ([]domain.Comment, int, error)
	GetTree(ctx context.Context, rootID int) ([]domain.Comment, error)
	Delete(ctx context.Context, id int) error
	Exists(ctx context.Context, id int) (bool, error)
	GetByID(ctx context.Context, id int) (domain.Comment, error)
}
