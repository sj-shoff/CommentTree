package domain

import "time"

type Comment struct {
	ID        int
	ParentID  *int
	Content   string
	Author    string
	CreatedAt time.Time
	UpdatedAt time.Time
	Children  []Comment
}

type CommentTree struct {
	Comments []Comment
	Total    int
	Page     int
	PageSize int
	HasNext  bool
	HasPrev  bool
}
