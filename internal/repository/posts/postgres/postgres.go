package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"comments-system/internal/domain"
	"comments-system/internal/usecase/posts"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

type PostsRepository struct {
	db      *dbpg.DB
	retries retry.Strategy
}

func NewPostsRepository(db *dbpg.DB, retries retry.Strategy) *PostsRepository {
	return &PostsRepository{
		db:      db,
		retries: retries,
	}
}

func (r *PostsRepository) Create(ctx context.Context, post domain.Post) (domain.Post, error) {
	var id int
	var createdAt, updatedAt time.Time
	query := `INSERT INTO posts (title, content, author, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id, created_at, updated_at`
	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, post.Title, post.Content, post.Author)
	if err != nil {
		return domain.Post{}, fmt.Errorf("failed to query row: %w", err)
	}
	err = row.Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return domain.Post{}, fmt.Errorf("failed to scan: %w", err)
	}
	post.ID = id
	post.CreatedAt = createdAt
	post.UpdatedAt = updatedAt
	return post, nil
}

func (r *PostsRepository) GetAll(ctx context.Context, page, pageSize int, searchQuery, sortBy, sortOrder string) ([]domain.Post, int, error) {
	where := ""
	params := []interface{}{}
	paramCount := 1
	if searchQuery != "" {
		where = "WHERE title ILIKE $1 OR content ILIKE $2"
		params = append(params, "%"+searchQuery+"%", "%"+searchQuery+"%")
		paramCount += 2
	}

	countQuery := `SELECT COUNT(*) FROM posts ` + where
	row, err := r.db.QueryRowWithRetry(ctx, r.retries, countQuery, params...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query row: %w", err)
	}
	var total int
	err = row.Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to scan: %w", err)
	}

	sortField := "created_at"
	switch sortBy {
	case "id":
		sortField = "id"
	case "title":
		sortField = "title"
	}
	sortDir := "DESC"
	if sortOrder == "asc" {
		sortDir = "ASC"
	}

	query := `SELECT id, title, content, author, created_at, updated_at FROM posts ` + where + ` ORDER BY ` + sortField + ` ` + sortDir + ` LIMIT $` + fmt.Sprint(paramCount) + ` OFFSET $` + fmt.Sprint(paramCount+1)
	params = append(params, pageSize, (page-1)*pageSize)

	rows, err := r.db.QueryWithRetry(ctx, r.retries, query, params...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query posts: %w", err)
	}
	defer rows.Close()

	var posts []domain.Post
	for rows.Next() {
		var p domain.Post
		err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.Author, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan post row: %w", err)
		}
		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating posts: %w", err)
	}

	return posts, total, nil
}

func (r *PostsRepository) GetByID(ctx context.Context, id int) (domain.Post, error) {
	var p domain.Post
	query := `SELECT id, title, content, author, created_at, updated_at FROM posts WHERE id = $1`
	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, id)
	if err != nil {
		return domain.Post{}, fmt.Errorf("failed to query row: %w", err)
	}
	err = row.Scan(&p.ID, &p.Title, &p.Content, &p.Author, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return domain.Post{}, posts.ErrPostNotFound
	}
	if err != nil {
		return domain.Post{}, fmt.Errorf("failed to scan: %w", err)
	}
	return p, nil
}

func (r *PostsRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM posts WHERE id = $1`
	_, err := r.db.ExecWithRetry(ctx, r.retries, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}
	return nil
}

func (r *PostsRepository) Exists(ctx context.Context, id int) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM posts WHERE id = $1`
	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, id)
	if err != nil {
		return false, fmt.Errorf("failed to query row: %w", err)
	}
	err = row.Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to scan: %w", err)
	}
	return count > 0, nil
}

func (r *PostsRepository) GetCommentsCount(ctx context.Context, postID int) (int, error) {
	query := `SELECT COUNT(*) FROM comments WHERE post_id = $1`
	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, postID)
	if err != nil {
		return 0, fmt.Errorf("failed to query row: %w", err)
	}
	var count int
	err = row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to scan: %w", err)
	}
	return count, nil
}
