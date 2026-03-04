package handler

import (
	"comment-tree/internal/domain"
	"comment-tree/internal/models"
	"comment-tree/internal/usecase"
	"errors"
	"net/http"
	"strconv"

	"github.com/wb-go/wbf/ginext"
)

// Handler — HTTP обработчик комментариев
type Handler struct {
	commentUC *usecase.CommentUseCase
}

// NewHandler создает новый обработчик
func NewHandler(commentUC *usecase.CommentUseCase) *Handler {
	return &Handler{
		commentUC: commentUC,
	}
}

// RegisterRoutes регистрирует все роуты
func (h *Handler) RegisterRoutes(router *ginext.Engine) {
	router.POST("/api/comments", h.CreateComment)
	router.GET("/api/comments", h.GetComments)
	router.GET("/api/comments/:id", h.GetComment)
	router.DELETE("/api/comments/:id", h.DeleteComment)
	router.POST("/api/comments/search", h.SearchComments)
}

// CreateComment создает новый комментарий
// POST /api/comments
func (h *Handler) CreateComment(c *ginext.Context) {
	var req models.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	comment, err := h.commentUC.CreateComment(c.Request.Context(), req.Author, req.Content, req.ParentID)
	if err != nil {
		if errors.Is(err, domain.ErrEmptyAuthor) || errors.Is(err, domain.ErrEmptyContent) {
			c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, comment)
}

// GetComments получает комментарии
// GET /api/comments?parent={id}&limit={limit}&offset={offset}
func (h *Handler) GetComments(c *ginext.Context) {
	parentIDStr := c.Query("parent")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	var parentID *int64
	if parentIDStr != "" {
		id, err := strconv.ParseInt(parentIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, ginext.H{"error": "invalid parent id"})
			return
		}
		parentID = &id
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	comments, err := h.commentUC.GetComments(c.Request.Context(), parentID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ginext.H{"comments": comments})
}

// GetComment получает комментарий по ID
// GET /api/comments/:id
func (h *Handler) GetComment(c *ginext.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "invalid comment id"})
		return
	}

	comment, err := h.commentUC.GetComment(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrCommentNotFound) {
			c.JSON(http.StatusNotFound, ginext.H{"error": "comment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, comment)
}

// DeleteComment удаляет комментарий
// DELETE /api/comments/:id
func (h *Handler) DeleteComment(c *ginext.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "invalid comment id"})
		return
	}

	err = h.commentUC.DeleteComment(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrCommentNotFound) {
			c.JSON(http.StatusNotFound, ginext.H{"error": "comment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ginext.H{"message": "comment deleted"})
}

// SearchComments ищет комментарии
// POST /api/comments/search
func (h *Handler) SearchComments(c *ginext.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	comments, err := h.commentUC.SearchComments(c.Request.Context(), req.Query, req.Limit, req.Offset)
	if err != nil {
		if errors.Is(err, domain.ErrEmptyContent) {
			c.JSON(http.StatusBadRequest, ginext.H{"error": "query is required"})
			return
		}
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ginext.H{"comments": comments})
}
