package comments

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"comments-system/internal/domain"
	"comments-system/internal/http-server/handler/comments/dto"
	"comments-system/internal/usecase/comments"

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
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	comment := domain.Comment{
		PostID:   req.PostID,
		ParentID: req.ParentID,
		Content:  req.Content,
		Author:   req.Author,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	createdComment, err := h.usecase.CreateComment(ctx, comment)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create comment")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := dto.CommentResponse{
		ID:        createdComment.ID,
		PostID:    createdComment.PostID,
		ParentID:  createdComment.ParentID,
		Content:   createdComment.Content,
		Author:    createdComment.Author,
		CreatedAt: createdComment.CreatedAt,
		UpdatedAt: createdComment.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *CommentsHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	var req dto.GetCommentsRequest

	postIDStr := r.URL.Query().Get("post_id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}
	req.PostID = postID

	parentIDStr := r.URL.Query().Get("parent")
	if parentIDStr != "" {
		parentID, err := strconv.Atoi(parentIDStr)
		if err != nil {
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

	tree, err := h.usecase.GetComments(ctx, req.PostID, req.ParentID, req.Page, req.PageSize, req.Search, req.SortBy, req.SortOrder)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get comments")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := dto.CommentsResponse{
		Comments: convertToResponse(tree.Comments),
		Total:    tree.Total,
		Page:     tree.Page,
		PageSize: tree.PageSize,
		HasNext:  tree.HasNext,
		HasPrev:  tree.HasPrev,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *CommentsHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	commentIDStr := chi.URLParam(r, "id")
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	err = h.usecase.DeleteComment(ctx, commentID)
	if err != nil {
		h.logger.Error().Err(err).Int("comment_id", commentID).Msg("Failed to delete comment")

		if err == comments.ErrCommentNotFound {
			http.Error(w, "Comment not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func convertToResponse(comments []domain.Comment) []dto.CommentResponse {
	responses := make([]dto.CommentResponse, len(comments))
	for i, comment := range comments {
		resp := dto.CommentResponse{
			ID:        comment.ID,
			PostID:    comment.PostID,
			ParentID:  comment.ParentID,
			Content:   comment.Content,
			Author:    comment.Author,
			CreatedAt: comment.CreatedAt,
			UpdatedAt: comment.UpdatedAt,
		}

		if len(comment.Children) > 0 {
			resp.Children = convertToResponse(comment.Children)
		}

		responses[i] = resp
	}
	return responses
}
