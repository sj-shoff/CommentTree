package comments_usecase

import (
	"context"
	"fmt"

	"comments-system/internal/domain"

	"github.com/wb-go/wbf/zlog"
)

type CommentsUsecase struct {
	repo   commentsRepo
	logger *zlog.Zerolog
}

func NewCommentsUsecase(repo commentsRepo, logger *zlog.Zerolog) *CommentsUsecase {
	return &CommentsUsecase{
		repo:   repo,
		logger: logger,
	}
}

func (u *CommentsUsecase) CreateComment(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
	if comment.Content == "" {
		return domain.Comment{}, ErrContentRequired
	}
	if comment.Author == "" {
		return domain.Comment{}, ErrAuthorRequired
	}
	if len(comment.Content) > 1000 {
		return domain.Comment{}, ErrContentTooLong
	}
	if len(comment.Author) > 50 {
		return domain.Comment{}, ErrAuthorTooLong
	}

	if comment.ParentID != nil {
		exists, err := u.repo.Exists(ctx, *comment.ParentID)
		if err != nil {
			return domain.Comment{}, err
		}
		if !exists {
			return domain.Comment{}, fmt.Errorf("%w: parent comment %d not found", ErrInvalidParentID, *comment.ParentID)
		}
	}

	createdComment, err := u.repo.Create(ctx, comment)
	if err != nil {
		return domain.Comment{}, err
	}

	return createdComment, nil
}

func (u *CommentsUsecase) GetComments(ctx context.Context, parentID *int, page, pageSize int, searchQuery, sortBy, sortOrder string) (domain.CommentTree, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	if parentID != nil {
		exists, err := u.repo.Exists(ctx, *parentID)
		if err != nil {
			return domain.CommentTree{}, err
		}
		if !exists {
			return domain.CommentTree{}, ErrCommentNotFound
		}
	}

	comments, total, err := u.repo.GetTree(ctx, parentID, page, pageSize, searchQuery, sortBy, sortOrder)
	if err != nil {
		return domain.CommentTree{}, err
	}

	return domain.CommentTree{
		Comments: comments,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasNext:  (page * pageSize) < total,
		HasPrev:  page > 1,
	}, nil
}

func (u *CommentsUsecase) DeleteComment(ctx context.Context, id int) error {
	if id <= 0 {
		return ErrInvalidCommentID
	}

	exists, err := u.repo.Exists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return ErrCommentNotFound
	}

	err = u.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
