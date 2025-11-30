package posts

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"comments-system/internal/domain"
	"comments-system/internal/http-server/handler/posts/dto"
	"comments-system/internal/usecase/posts"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/wb-go/wbf/zlog"
)

type PostsHandler struct {
	usecase  postsUsecase
	logger   *zlog.Zerolog
	validate *validator.Validate
}

func NewPostsHandler(usecase postsUsecase, logger *zlog.Zerolog) *PostsHandler {
	return &PostsHandler{
		usecase:  usecase,
		logger:   logger,
		validate: validator.New(),
	}
}

func (h *PostsHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var req dto.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	post := domain.Post{
		Title:   req.Title,
		Content: req.Content,
		Author:  req.Author,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	createdPost, err := h.usecase.CreatePost(ctx, post)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create post")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := dto.PostResponse{
		ID:            createdPost.ID,
		Title:         createdPost.Title,
		Content:       createdPost.Content,
		Author:        createdPost.Author,
		CreatedAt:     createdPost.CreatedAt,
		UpdatedAt:     createdPost.UpdatedAt,
		CommentsCount: createdPost.CommentsCount,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *PostsHandler) GetPosts(w http.ResponseWriter, r *http.Request) {
	var req dto.GetPostsRequest

	req.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	req.PageSize, _ = strconv.Atoi(r.URL.Query().Get("page_size"))
	req.Search = r.URL.Query().Get("search")
	req.SortBy = r.URL.Query().Get("sort_by")
	req.SortOrder = r.URL.Query().Get("sort_order")

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	tree, err := h.usecase.GetPosts(ctx, req.Page, req.PageSize, req.Search, req.SortBy, req.SortOrder)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get posts")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := dto.PostsResponse{
		Posts:    convertToResponse(tree.Posts),
		Total:    tree.Total,
		Page:     tree.Page,
		PageSize: tree.PageSize,
		HasNext:  tree.HasNext,
		HasPrev:  tree.HasPrev,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *PostsHandler) GetPostByID(w http.ResponseWriter, r *http.Request) {
	postIDStr := chi.URLParam(r, "id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	post, err := h.usecase.GetPostByID(ctx, postID)
	if err != nil {
		h.logger.Error().Err(err).Int("post_id", postID).Msg("Failed to get post")

		if err == posts.ErrPostNotFound {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := dto.PostResponse{
		ID:            post.ID,
		Title:         post.Title,
		Content:       post.Content,
		Author:        post.Author,
		CreatedAt:     post.CreatedAt,
		UpdatedAt:     post.UpdatedAt,
		CommentsCount: post.CommentsCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *PostsHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	postIDStr := chi.URLParam(r, "id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	err = h.usecase.DeletePost(ctx, postID)
	if err != nil {
		h.logger.Error().Err(err).Int("post_id", postID).Msg("Failed to delete post")

		if err == posts.ErrPostNotFound {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func convertToResponse(posts []domain.Post) []dto.PostResponse {
	responses := make([]dto.PostResponse, len(posts))
	for i, post := range posts {
		responses[i] = dto.PostResponse{
			ID:            post.ID,
			Title:         post.Title,
			Content:       post.Content,
			Author:        post.Author,
			CreatedAt:     post.CreatedAt,
			UpdatedAt:     post.UpdatedAt,
			CommentsCount: post.CommentsCount,
		}
	}
	return responses
}
