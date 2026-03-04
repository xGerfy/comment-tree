package domain

import "context"

// CommentService — интерфейс сервиса комментариев
type CommentService interface {
	CreateComment(ctx context.Context, author, content string, parentID *int64) (*Comment, error)
	GetComments(ctx context.Context, parentID *int64, limit, offset int) ([]Comment, error)
	GetComment(ctx context.Context, id int64) (*Comment, error)
	DeleteComment(ctx context.Context, id int64) error
	SearchComments(ctx context.Context, query string, limit, offset int) ([]Comment, error)
}
