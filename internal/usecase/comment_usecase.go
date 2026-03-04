package usecase

import (
	"comment-tree/internal/domain"
	"context"
	"errors"
)

// Константы валидации
const (
	MaxContentLength = 10000
	MaxAuthorLength  = 255
)

// CommentUseCase — реализация сервиса комментариев
type CommentUseCase struct {
	repo domain.CommentRepository
}

// NewCommentUseCase создает новый use case
func NewCommentUseCase(repo domain.CommentRepository) *CommentUseCase {
	return &CommentUseCase{repo: repo}
}

// CreateComment создает новый комментарий
func (uc *CommentUseCase) CreateComment(ctx context.Context, author, content string, parentID *int64) (*domain.Comment, error) {
	if author == "" {
		return nil, domain.ErrEmptyAuthor
	}
	if len(author) > MaxAuthorLength {
		return nil, errors.New("author name too long (max 255 characters)")
	}
	if content == "" {
		return nil, domain.ErrEmptyContent
	}
	if len(content) > MaxContentLength {
		return nil, errors.New("content too long (max 10000 characters)")
	}

	comment := &domain.Comment{
		Author:  author,
		Content: content,
	}

	if parentID != nil && *parentID > 0 {
		comment.ParentID = parentID
	}

	if err := uc.repo.Create(ctx, comment); err != nil {
		return nil, err
	}

	return comment, nil
}

// GetComments получает комментарии
func (uc *CommentUseCase) GetComments(ctx context.Context, parentID *int64, limit, offset int) ([]domain.Comment, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return uc.repo.GetTree(ctx, parentID, limit, offset)
}

// GetComment получает комментарий по ID
func (uc *CommentUseCase) GetComment(ctx context.Context, id int64) (*domain.Comment, error) {
	if id <= 0 {
		return nil, domain.ErrCommentNotFound
	}

	return uc.repo.GetByID(ctx, id)
}

// DeleteComment удаляет комментарий
func (uc *CommentUseCase) DeleteComment(ctx context.Context, id int64) error {
	if id <= 0 {
		return domain.ErrCommentNotFound
	}

	return uc.repo.Delete(ctx, id)
}

// SearchComments ищет комментарии
func (uc *CommentUseCase) SearchComments(ctx context.Context, query string, limit, offset int) ([]domain.Comment, error) {
	if query == "" {
		return nil, domain.ErrEmptyContent
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return uc.repo.Search(ctx, query, limit, offset)
}
