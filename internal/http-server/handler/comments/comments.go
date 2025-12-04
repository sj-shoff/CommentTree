package comments

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"comments-system/internal/domain"
	"comments-system/internal/http-server/handler/comments/dto"
	comments_usecase "comments-system/internal/usecase/comments"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/wb-go/wbf/zlog"
)

type CommentsHandler struct {
	usecase  commentsUsecase
	logger   *zlog.Zerolog
	validate *validator.Validate
}

func NewCommentsHandler(usecase commentsUsecase, logger *zlog.Zerolog) *CommentsHandler {
	return &CommentsHandler{
		usecase:  usecase,
		logger:   logger,
		validate: validator.New(),
	}
}

func (h *CommentsHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.logger.Error().Err(err).Msg("Request validation failed")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	comment := domain.Comment{
		ParentID: req.ParentID,
		Content:  req.Content,
		Author:   req.Author,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	createdComment, err := h.usecase.CreateComment(ctx, comment)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create comment")

		if errors.Is(err, comments_usecase.ErrInvalidParentID) {
			http.Error(w, "Parent comment not found", http.StatusBadRequest)
			return
		}
		if errors.Is(err, comments_usecase.ErrContentRequired) ||
			errors.Is(err, comments_usecase.ErrAuthorRequired) ||
			errors.Is(err, comments_usecase.ErrContentTooLong) ||
			errors.Is(err, comments_usecase.ErrAuthorTooLong) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := dto.CommentResponse{
		ID:        createdComment.ID,
		ParentID:  createdComment.ParentID,
		Content:   createdComment.Content,
		Author:    createdComment.Author,
		CreatedAt: createdComment.CreatedAt,
		UpdatedAt: createdComment.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode response")
	}
}

func (h *CommentsHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	var req dto.GetCommentsRequest

	parentIDStr := r.URL.Query().Get("parent")
	if parentIDStr != "" {
		parentID, err := strconv.Atoi(parentIDStr)
		if err != nil {
			h.logger.Error().Err(err).Str("parent_id", parentIDStr).Msg("Invalid parent ID")
			http.Error(w, "Invalid parent ID", http.StatusBadRequest)
			return
		}
		req.ParentID = &parentID
	}

	req.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	req.PageSize, _ = strconv.Atoi(r.URL.Query().Get("page_size"))
	req.Search = r.URL.Query().Get("search")
	req.SortBy = r.URL.Query().Get("sort_by")
	req.SortOrder = r.URL.Query().Get("sort_order")

	if err := req.Validate(); err != nil {
		h.logger.Error().Err(err).Msg("Invalid query parameters")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	tree, err := h.usecase.GetComments(ctx, req.ParentID, req.Page, req.PageSize, req.Search, req.SortBy, req.SortOrder)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get comments")

		if errors.Is(err, comments_usecase.ErrCommentNotFound) {
			http.Error(w, "Comment not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := dto.FromDomainCommentTree(tree)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode response")
	}
}

func (h *CommentsHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	commentIDStr := chi.URLParam(r, "id")
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		h.logger.Error().Err(err).Str("comment_id", commentIDStr).Msg("Invalid comment ID")
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	err = h.usecase.DeleteComment(ctx, commentID)
	if err != nil {
		h.logger.Error().Err(err).Int("comment_id", commentID).Msg("Failed to delete comment")

		if errors.Is(err, comments_usecase.ErrCommentNotFound) {
			http.Error(w, "Comment not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, comments_usecase.ErrInvalidCommentID) {
			http.Error(w, "Invalid comment ID", http.StatusBadRequest)
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
