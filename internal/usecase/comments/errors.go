package comments

import "errors"

var (
	ErrInvalidCommentID = errors.New("invalid comment ID")
	ErrCommentNotFound  = errors.New("comment not found")
	ErrInvalidParentID  = errors.New("invalid parent ID")
	ErrInvalidPostID    = errors.New("invalid post ID")
	ErrContentRequired  = errors.New("content is required")
	ErrAuthorRequired   = errors.New("author is required")
	ErrContentTooLong   = errors.New("content is too long")
	ErrAuthorTooLong    = errors.New("author is too long")
)
