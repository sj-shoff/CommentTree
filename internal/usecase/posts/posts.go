package posts

import (
	"context"
	"time"

	"comments-system/internal/domain"

	"github.com/wb-go/wbf/zlog"
)

type PostsUsecase struct {
	repo   postsRepo
	cache  postsCache
	logger *zlog.Zerolog
}

func NewPostsUsecase(repo postsRepo, cache postsCache, logger *zlog.Zerolog) *PostsUsecase {
	return &PostsUsecase{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

func (u *PostsUsecase) CreatePost(ctx context.Context, post domain.Post) (domain.Post, error) {
	if post.Title == "" {
		return domain.Post{}, ErrTitleRequired
	}
	if post.Content == "" {
		return domain.Post{}, ErrContentRequired
	}
	if post.Author == "" {
		return domain.Post{}, ErrAuthorRequired
	}
	if len(post.Title) > 200 {
		return domain.Post{}, ErrTitleTooLong
	}
	if len(post.Content) > 10000 {
		return domain.Post{}, ErrContentTooLong
	}
	if len(post.Author) > 50 {
		return domain.Post{}, ErrAuthorTooLong
	}

	now := time.Now()
	post.CreatedAt = now
	post.UpdatedAt = now

	createdPost, err := u.repo.Create(ctx, post)
	if err != nil {
		return domain.Post{}, err
	}

	if err := u.cache.InvalidatePosts(ctx); err != nil {
		u.logger.Warn().Err(err).Msg("Failed to invalidate posts cache")
	}

	return createdPost, nil
}

// УПРОЩАЕМ - возвращаем просто слайс постов и общее количество
func (u *PostsUsecase) GetPosts(ctx context.Context, page, pageSize int, searchQuery, sortBy, sortOrder string) ([]domain.Post, int, error) {
	cachedPosts, totalCount, err := u.cache.GetPosts(ctx, page, pageSize, searchQuery, sortBy, sortOrder)
	if err == nil {
		return cachedPosts, totalCount, nil
	}

	u.logger.Warn().Err(err).Msg("Cache miss for posts, querying DB")

	posts, totalCount, err := u.repo.GetAll(ctx, page, pageSize, searchQuery, sortBy, sortOrder)
	if err != nil {
		return nil, 0, err
	}

	for i := range posts {
		count, countErr := u.repo.GetCommentsCount(ctx, posts[i].ID)
		if countErr != nil {
			u.logger.Warn().Err(countErr).Int("post_id", posts[i].ID).Msg("Failed to get comments count")
			count = 0
		}
		posts[i].CommentsCount = count
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := u.cache.SetPosts(ctx, page, pageSize, searchQuery, sortBy, sortOrder, posts, totalCount); err != nil {
			u.logger.Warn().Err(err).Msg("Failed to cache posts")
		}
	}()

	return posts, totalCount, nil
}

func (u *PostsUsecase) GetPostByID(ctx context.Context, id int) (domain.Post, error) {
	cachedPost, err := u.cache.GetPost(ctx, id)
	if err == nil {
		return cachedPost, nil
	}

	u.logger.Warn().Err(err).Int("post_id", id).Msg("Cache miss for post, querying DB")

	post, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Post{}, err
	}

	count, countErr := u.repo.GetCommentsCount(ctx, id)
	if countErr != nil {
		u.logger.Warn().Err(countErr).Int("post_id", id).Msg("Failed to get comments count")
		count = 0
	}
	post.CommentsCount = count

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := u.cache.SetPost(ctx, post); err != nil {
			u.logger.Warn().Err(err).Int("post_id", id).Msg("Failed to cache post")
		}
	}()

	return post, nil
}

func (u *PostsUsecase) DeletePost(ctx context.Context, id int) error {
	if id <= 0 {
		return ErrInvalidPostID
	}

	exists, err := u.repo.Exists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return ErrPostNotFound
	}

	if err := u.repo.Delete(ctx, id); err != nil {
		return err
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		u.cache.InvalidatePost(ctx, id)
		u.cache.InvalidatePosts(ctx)
	}()

	return nil
}
