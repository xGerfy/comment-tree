package domain

import "context"

// CommentRepository — интерфейс репозитория комментариев
type CommentRepository interface {
	Create(ctx context.Context, comment *Comment) error
	GetByID(ctx context.Context, id int64) (*Comment, error)
	GetRootComments(ctx context.Context, limit, offset int) ([]Comment, error)
	GetChildren(ctx context.Context, parentID int64) ([]Comment, error)
	GetTree(ctx context.Context, parentID *int64, limit, offset int) ([]Comment, error)
	Delete(ctx context.Context, id int64) error
	Search(ctx context.Context, query string, limit, offset int) ([]Comment, error)
}
