package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"comments-system/internal/domain"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

type CommentsRepository struct {
	db      *dbpg.DB
	retries retry.Strategy
}

func NewCommentsRepository(db *dbpg.DB, retries retry.Strategy) *CommentsRepository {
	return &CommentsRepository{
		db:      db,
		retries: retries,
	}
}

func (r *CommentsRepository) Create(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
	var id int
	var createdAt, updatedAt time.Time

	query := `INSERT INTO comments (parent_id, content, author, created_at, updated_at) 
	          VALUES ($1, $2, $3, NOW(), NOW()) 
	          RETURNING id, created_at, updated_at`

	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, comment.ParentID, comment.Content, comment.Author)
	if err != nil {
		return domain.Comment{}, fmt.Errorf("failed to create comment: %w", err)
	}

	err = row.Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return domain.Comment{}, fmt.Errorf("failed to scan created comment: %w", err)
	}

	comment.ID = id
	comment.CreatedAt = createdAt
	comment.UpdatedAt = updatedAt
	return comment, nil
}

func (r *CommentsRepository) GetByParent(ctx context.Context, parentID *int, page, pageSize int, searchQuery, sortBy, sortOrder string) ([]domain.Comment, int, error) {
	var whereConditions []string
	var params []interface{}

	if parentID != nil {
		whereConditions = append(whereConditions, "parent_id = $1")
		params = append(params, *parentID)
	} else {
		whereConditions = append(whereConditions, "parent_id IS NULL")
	}

	if searchQuery != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("content ILIKE $%d", len(params)+1))
		params = append(params, "%"+searchQuery+"%")
	}

	whereClause := "WHERE " + strings.Join(whereConditions, " AND ")

	countQuery := `SELECT COUNT(*) FROM comments ` + whereClause
	row, err := r.db.QueryRowWithRetry(ctx, r.retries, countQuery, params...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count comments: %w", err)
	}

	var total int
	err = row.Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to scan count: %w", err)
	}

	sortField := "created_at"
	switch sortBy {
	case "id":
		sortField = "id"
	case "updated_at":
		sortField = "updated_at"
	}

	sortDir := "DESC"
	if sortOrder == "asc" {
		sortDir = "ASC"
	}

	query := `SELECT id, parent_id, content, author, created_at, updated_at 
	          FROM comments ` + whereClause +
		` ORDER BY ` + sortField + ` ` + sortDir +
		` LIMIT $` + strconv.Itoa(len(params)+1) +
		` OFFSET $` + strconv.Itoa(len(params)+2)

	params = append(params, pageSize, (page-1)*pageSize)

	rows, err := r.db.QueryWithRetry(ctx, r.retries, query, params...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query comments: %w", err)
	}
	defer rows.Close()

	var comments []domain.Comment
	for rows.Next() {
		var c domain.Comment
		var pid sql.NullInt32

		err := rows.Scan(&c.ID, &pid, &c.Content, &c.Author, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan comment row: %w", err)
		}

		if pid.Valid {
			pidInt := int(pid.Int32)
			c.ParentID = &pidInt
		}

		comments = append(comments, c)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating comments: %w", err)
	}

	return comments, total, nil
}

func (r *CommentsRepository) GetTree(ctx context.Context, rootID int) ([]domain.Comment, error) {
	query := `
	WITH RECURSIVE comment_tree AS (
	  SELECT id, parent_id, content, author, created_at, updated_at 
	  FROM comments 
	  WHERE id = $1
	  
	  UNION ALL
	  
	  SELECT c.id, c.parent_id, c.content, c.author, c.created_at, c.updated_at 
	  FROM comments c 
	  INNER JOIN comment_tree ct ON c.parent_id = ct.id
	)
	SELECT id, parent_id, content, author, created_at, updated_at 
	FROM comment_tree
	ORDER BY created_at ASC
	`

	rows, err := r.db.QueryWithRetry(ctx, r.retries, query, rootID)
	if err != nil {
		return nil, fmt.Errorf("failed to query comment tree: %w", err)
	}
	defer rows.Close()

	var comments []domain.Comment
	for rows.Next() {
		var c domain.Comment
		var pid sql.NullInt32

		err := rows.Scan(&c.ID, &pid, &c.Content, &c.Author, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment row: %w", err)
		}

		if pid.Valid {
			pidInt := int(pid.Int32)
			c.ParentID = &pidInt
		}

		comments = append(comments, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating comment tree: %w", err)
	}

	return comments, nil
}

func (r *CommentsRepository) Exists(ctx context.Context, id int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM comments WHERE id = $1)`

	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, id)
	if err != nil {
		return false, fmt.Errorf("failed to check comment existence: %w", err)
	}

	var exists bool
	err = row.Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to scan existence: %w", err)
	}

	return exists, nil
}

func (r *CommentsRepository) GetByID(ctx context.Context, id int) (domain.Comment, error) {
	var c domain.Comment
	var pid sql.NullInt32

	query := `SELECT id, parent_id, content, author, created_at, updated_at FROM comments WHERE id = $1`

	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, id)
	if err != nil {
		return domain.Comment{}, fmt.Errorf("failed to query comment: %w", err)
	}

	err = row.Scan(&c.ID, &pid, &c.Content, &c.Author, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return domain.Comment{}, comments_usecase.ErrCommentNotFound
	}
	if err != nil {
		return domain.Comment{}, fmt.Errorf("failed to scan comment: %w", err)
	}

	if pid.Valid {
		pidInt := int(pid.Int32)
		c.ParentID = &pidInt
	}

	return c, nil
}

func (r *CommentsRepository) Delete(ctx context.Context, id int) error {
	query := `
	WITH RECURSIVE descendants AS (
	  SELECT id FROM comments WHERE id = $1
	  
	  UNION
	  
	  SELECT c.id FROM comments c 
	  INNER JOIN descendants d ON c.parent_id = d.id
	)
	DELETE FROM comments WHERE id IN (SELECT id FROM descendants)
	`

	_, err := r.db.ExecWithRetry(ctx, r.retries, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete comment tree: %w", err)
	}

	return nil
}
