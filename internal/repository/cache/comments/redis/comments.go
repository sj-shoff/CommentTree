package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"comments-system/internal/config"
	"comments-system/internal/domain"
	"comments-system/internal/repository/cache"

	wbfredis "github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/retry"
)

type сommentsCache struct {
	client  *wbfredis.Client
	retries retry.Strategy
}

func NewCommentsCache(cfg *config.Config, retries retry.Strategy) *сommentsCache {
	client := wbfredis.New(cfg.RedisAddr(), cfg.Redis.Pass, cfg.Redis.DB)
	return &сommentsCache{
		client:  client,
		retries: retries,
	}
}

func (c *сommentsCache) GetCommentTree(ctx context.Context, rootID int) ([]domain.Comment, error) {
	key := fmt.Sprintf("comment:tree:%d", rootID)

	data, err := c.client.GetWithRetry(ctx, c.retries, key)
	if err != nil {
		if err.Error() == "redis: nil" {
			return nil, fmt.Errorf("%w: key %s not found", cache.ErrCacheMiss, key)
		}
		return nil, fmt.Errorf("failed to get comment tree from cache: %w", err)
	}

	var comments []domain.Comment
	if err := json.Unmarshal([]byte(data), &comments); err != nil {
		return nil, fmt.Errorf("failed to unmarshal comment tree: %w", err)
	}

	return comments, nil
}

func (c *сommentsCache) SetCommentTree(ctx context.Context, rootID int, comments []domain.Comment) error {
	key := fmt.Sprintf("comment:tree:%d", rootID)

	data, err := json.Marshal(comments)
	if err != nil {
		return fmt.Errorf("failed to marshal comment tree: %w", err)
	}

	err = c.client.SetWithRetry(ctx, c.retries, key, string(data))
	if err != nil {
		return fmt.Errorf("failed to set comment tree in cache: %w", err)
	}
	return nil
}

func (c *сommentsCache) InvalidateCommentTree(ctx context.Context, rootID int) error {
	key := fmt.Sprintf("comment:tree:%d", rootID)

	err := c.client.DelWithRetry(ctx, c.retries, key)
	if err != nil {
		return fmt.Errorf("failed to invalidate comment tree: %w", err)
	}
	return nil
}

func (c *сommentsCache) GetComments(ctx context.Context, parentID *int, page, pageSize int, searchQuery, sortBy, sortOrder string) ([]domain.Comment, int, error) {
	key := c.getCommentsCacheKey(parentID, page, pageSize, searchQuery, sortBy, sortOrder)

	data, err := c.client.GetWithRetry(ctx, c.retries, key)
	if err != nil {
		if err.Error() == "redis: nil" {
			return nil, 0, fmt.Errorf("%w: key %s not found", cache.ErrCacheMiss, key)
		}
		return nil, 0, fmt.Errorf("failed to get comments from cache: %w", err)
	}

	var cachedData struct {
		Comments   []domain.Comment
		TotalCount int
	}

	if err := json.Unmarshal([]byte(data), &cachedData); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal comments: %w", err)
	}

	return cachedData.Comments, cachedData.TotalCount, nil
}

func (c *сommentsCache) SetComments(ctx context.Context, parentID *int, page, pageSize int, searchQuery, sortBy, sortOrder string, comments []domain.Comment, totalCount int) error {
	key := c.getCommentsCacheKey(parentID, page, pageSize, searchQuery, sortBy, sortOrder)

	data, err := json.Marshal(struct {
		Comments   []domain.Comment
		TotalCount int
	}{
		Comments:   comments,
		TotalCount: totalCount,
	})

	if err != nil {
		return fmt.Errorf("failed to marshal comments: %w", err)
	}

	err = c.client.SetWithRetry(ctx, c.retries, key, string(data))
	if err != nil {
		return fmt.Errorf("failed to set comments in cache: %w", err)
	}
	return nil
}

func (c *сommentsCache) InvalidateComments(ctx context.Context, parentID *int) error {
	if parentID != nil {
		key := fmt.Sprintf("comment:tree:%d", *parentID)
		err := c.client.DelWithRetry(ctx, c.retries, key)
		if err != nil {
			return fmt.Errorf("failed to invalidate comment tree: %w", err)
		}
	}

	rootKey := "comment:tree:root"
	err := c.client.DelWithRetry(ctx, c.retries, rootKey)
	if err != nil {
		return fmt.Errorf("failed to invalidate root comments: %w", err)
	}

	return nil
}

func (c *сommentsCache) getCommentsCacheKey(parentID *int, page, pageSize int, searchQuery, sortBy, sortOrder string) string {
	if parentID != nil {
		return fmt.Sprintf("comments:parent:%d:page:%d:size:%d:search:%s:sort:%s:%s",
			*parentID, page, pageSize, searchQuery, sortBy, sortOrder)
	}
	return fmt.Sprintf("comments:root:page:%d:size:%d:search:%s:sort:%s:%s",
		page, pageSize, searchQuery, sortBy, sortOrder)
}
