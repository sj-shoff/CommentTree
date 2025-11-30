package comments

import "errors"

var (
	ErrInvalidCommentID  = errors.New("invalid comment ID")
	ErrCommentNotFound   = errors.New("comment not found")
	ErrInvalidParentID   = errors.New("invalid parent ID")
	ErrInvalidPagination = errors.New("invalid pagination parameters")
)
