package domain

import (
	"errors"
	"time"
)

// Comment — сущность комментария
type Comment struct {
	ID        int64     `json:"id"`
	ParentID  *int64    `json:"parent_id,omitempty"`
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Children  []Comment `json:"children,omitempty"`
}

// Ошибки доменной области
var (
	ErrCommentNotFound = errors.New("comment not found")
	ErrInvalidParentID = errors.New("invalid parent id")
	ErrEmptyContent    = errors.New("content is required")
	ErrEmptyAuthor     = errors.New("author is required")
)
