package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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

	query := `INSERT INTO posts (title, content, author, created_at, updated_at) 
              VALUES ($1, $2, $3, NOW(), NOW()) 
              RETURNING id, created_at, updated_at`

	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, post.Title, post.Content, post.Author)
	if err != nil {
		return domain.Post{}, fmt.Errorf("failed to create post: %w", err)
	}

	err = row.Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return domain.Post{}, fmt.Errorf("failed to scan created post: %w", err)
	}

	post.ID = id
	post.CreatedAt = createdAt
	post.UpdatedAt = updatedAt
	return post, nil
}

func (r *PostsRepository) GetAll(ctx context.Context, page, pageSize int, searchQuery, sortBy, sortOrder string) ([]domain.Post, int, error) {
	whereConditions := []string{"1=1"}
	params := []interface{}{}
	paramCount := 0

	if searchQuery != "" {
		paramCount++
		whereConditions = append(whereConditions, fmt.Sprintf("(title ILIKE $%d OR content ILIKE $%d)", paramCount, paramCount))
		params = append(params, "%"+searchQuery+"%")
	}

	whereClause := strings.Join(whereConditions, " AND ")

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM posts WHERE %s`, whereClause)

	row, err := r.db.QueryRowWithRetry(ctx, r.retries, countQuery, params...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count posts: %w", err)
	}

	var total int
	if err := row.Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to scan count: %w", err)
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

	query := fmt.Sprintf(`
        SELECT id, title, content, author, created_at, updated_at 
        FROM posts 
        WHERE %s 
        ORDER BY %s %s 
        LIMIT $%d OFFSET $%d`,
		whereClause, sortField, sortDir, len(params)+1, len(params)+2)

	params = append(params, pageSize, (page-1)*pageSize)

	rows, err := r.db.QueryWithRetry(ctx, r.retries, query, params...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query posts: %w", err)
	}
	defer rows.Close()

	var posts []domain.Post
	for rows.Next() {
		var post domain.Post
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Author, &post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating posts: %w", err)
	}

	return posts, total, nil
}

func (r *PostsRepository) GetByID(ctx context.Context, id int) (domain.Post, error) {
	query := `SELECT id, title, content, author, created_at, updated_at FROM posts WHERE id = $1`

	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, id)
	if err != nil {
		return domain.Post{}, fmt.Errorf("failed to query post: %w", err)
	}

	var post domain.Post
	err = row.Scan(&post.ID, &post.Title, &post.Content, &post.Author, &post.CreatedAt, &post.UpdatedAt)
	if err == sql.ErrNoRows {
		return domain.Post{}, posts.ErrPostNotFound
	}
	if err != nil {
		return domain.Post{}, fmt.Errorf("failed to scan post: %w", err)
	}

	return post, nil
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
	query := `SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)`

	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, id)
	if err != nil {
		return false, fmt.Errorf("failed to check post existence: %w", err)
	}

	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to scan existence: %w", err)
	}

	return exists, nil
}

func (r *PostsRepository) GetCommentsCount(ctx context.Context, postID int) (int, error) {
	query := `SELECT COUNT(*) FROM comments WHERE post_id = $1`

	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, postID)
	if err != nil {
		return 0, fmt.Errorf("failed to count comments: %w", err)
	}

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to scan comments count: %w", err)
	}

	return count, nil
}
