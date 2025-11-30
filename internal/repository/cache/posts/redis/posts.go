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

type postsCache struct {
	client  *wbfredis.Client
	retries retry.Strategy
}

func NewPostsCache(cfg *config.Config, retries retry.Strategy) *postsCache {
	client := wbfredis.New(cfg.RedisAddr(), cfg.Redis.Pass, cfg.Redis.DB)
	return &postsCache{
		client:  client,
		retries: retries,
	}
}

func (c *postsCache) GetPost(ctx context.Context, id int) (domain.Post, error) {
	key := fmt.Sprintf("post:%d", id)

	data, err := c.client.GetWithRetry(ctx, c.retries, key)
	if err != nil {
		if err.Error() == "redis: nil" {
			return domain.Post{}, fmt.Errorf("%w: key %s not found", cache.ErrCacheMiss, key)
		}
		return domain.Post{}, fmt.Errorf("failed to get post from cache: %w", err)
	}

	var post domain.Post
	if err := json.Unmarshal([]byte(data), &post); err != nil {
		return domain.Post{}, fmt.Errorf("failed to unmarshal post: %w", err)
	}

	return post, nil
}

func (c *postsCache) SetPost(ctx context.Context, post domain.Post) error {
	key := fmt.Sprintf("post:%d", post.ID)

	data, err := json.Marshal(post)
	if err != nil {
		return fmt.Errorf("failed to marshal post: %w", err)
	}

	err = c.client.SetWithRetry(ctx, c.retries, key, string(data))
	if err != nil {
		return fmt.Errorf("failed to set post in cache: %w", err)
	}
	return nil
}

func (c *postsCache) InvalidatePost(ctx context.Context, id int) error {
	key := fmt.Sprintf("post:%d", id)

	err := c.client.DelWithRetry(ctx, c.retries, key)
	if err != nil {
		return fmt.Errorf("failed to invalidate post: %w", err)
	}
	return nil
}

func (c *postsCache) GetPosts(ctx context.Context, page, pageSize int, searchQuery, sortBy, sortOrder string) ([]domain.Post, int, error) {
	key := c.getPostsCacheKey(page, pageSize, searchQuery, sortBy, sortOrder)

	data, err := c.client.GetWithRetry(ctx, c.retries, key)
	if err != nil {
		if err.Error() == "redis: nil" {
			return nil, 0, fmt.Errorf("%w: key %s not found", cache.ErrCacheMiss, key)
		}
		return nil, 0, fmt.Errorf("failed to get posts from cache: %w", err)
	}

	var cachedData struct {
		Posts      []domain.Post
		TotalCount int
	}

	if err := json.Unmarshal([]byte(data), &cachedData); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal posts: %w", err)
	}

	return cachedData.Posts, cachedData.TotalCount, nil
}

func (c *postsCache) SetPosts(ctx context.Context, page, pageSize int, searchQuery, sortBy, sortOrder string, posts []domain.Post, totalCount int) error {
	key := c.getPostsCacheKey(page, pageSize, searchQuery, sortBy, sortOrder)

	data, err := json.Marshal(struct {
		Posts      []domain.Post
		TotalCount int
	}{
		Posts:      posts,
		TotalCount: totalCount,
	})

	if err != nil {
		return fmt.Errorf("failed to marshal posts: %w", err)
	}

	err = c.client.SetWithRetry(ctx, c.retries, key, string(data))
	if err != nil {
		return fmt.Errorf("failed to set posts in cache: %w", err)
	}
	return nil
}

func (c *postsCache) InvalidatePosts(ctx context.Context) error {
	if err := c.client.DelWithRetry(ctx, c.retries, "posts:list"); err != nil {
		return fmt.Errorf("failed to invalidate posts list cache: %w", err)
	}

	if err := c.client.DelWithRetry(ctx, c.retries, "posts:root"); err != nil {
		return fmt.Errorf("failed to invalidate root posts cache: %w", err)
	}

	return nil
}

func (c *postsCache) GetCommentsCount(ctx context.Context, postID int) (int, error) {
	key := fmt.Sprintf("post:%d:comments_count", postID)

	data, err := c.client.GetWithRetry(ctx, c.retries, key)
	if err != nil {
		if err.Error() == "redis: nil" {
			return 0, fmt.Errorf("%w: key %s not found", cache.ErrCacheMiss, key)
		}
		return 0, fmt.Errorf("failed to get comments count from cache: %w", err)
	}

	var count int
	if err := json.Unmarshal([]byte(data), &count); err != nil {
		return 0, fmt.Errorf("failed to unmarshal comments count: %w", err)
	}

	return count, nil
}

func (c *postsCache) SetCommentsCount(ctx context.Context, postID int, count int) error {
	key := fmt.Sprintf("post:%d:comments_count", postID)

	data, err := json.Marshal(count)
	if err != nil {
		return fmt.Errorf("failed to marshal comments count: %w", err)
	}

	err = c.client.SetWithRetry(ctx, c.retries, key, string(data))
	if err != nil {
		return fmt.Errorf("failed to set comments count in cache: %w", err)
	}
	return nil
}

func (c *postsCache) getPostsCacheKey(page, pageSize int, searchQuery, sortBy, sortOrder string) string {
	return fmt.Sprintf("posts:page:%d:size:%d:search:%s:sort:%s:%s",
		page, pageSize, searchQuery, sortBy, sortOrder)
}
