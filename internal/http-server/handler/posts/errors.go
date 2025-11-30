package posts

import "errors"

var (
	ErrInvalidPostID   = errors.New("invalid post ID")
	ErrPostNotFound    = errors.New("post not found")
	ErrTitleRequired   = errors.New("title is required")
	ErrContentRequired = errors.New("content is required")
	ErrAuthorRequired  = errors.New("author is required")
	ErrTitleTooLong    = errors.New("title is too long")
	ErrContentTooLong  = errors.New("content is too long")
	ErrAuthorTooLong   = errors.New("author name is too long")
)
