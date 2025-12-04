package comments

import (
	"comments-system/internal/domain"
	"context"
)

type commentsUsecase interface {
	CreateComment(ctx context.Context, comment domain.Comment) (domain.Comment, error)
	GetComments(ctx context.Context, parentID *int, page, pageSize int, searchQuery, sortBy, sortOrder string) (domain.CommentTree, error)
	DeleteComment(ctx context.Context, id int) error
}
