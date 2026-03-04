package postgres

import (
	"comment-tree/internal/domain"
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Таймауты для операций с БД
const (
	DBQueryTimeout = 5 * time.Second
	DBWriteTimeout = 10 * time.Second
)

// CommentRepository — реализация репозитория комментариев
type CommentRepository struct {
	db *sql.DB
}

// NewCommentRepository создает новый репозиторий
func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

// Create создает новый комментарий
func (r *CommentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	ctx, cancel := context.WithTimeout(ctx, DBWriteTimeout)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	path, err := r.buildPath(ctx, tx, comment.ParentID)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO comments (parent_id, author, content, path)
		VALUES (NULLIF($1, 0), $2, $3, $4::ltree)
		RETURNING id, parent_id, author, content, path, created_at, updated_at`

	var parentID sql.NullInt64
	if comment.ParentID != nil {
		parentID.Int64 = *comment.ParentID
		parentID.Valid = true
	}

	err = tx.QueryRowContext(ctx, query, parentID.Int64, comment.Author, comment.Content, path).Scan(
		&comment.ID,
		&parentID,
		&comment.Author,
		&comment.Content,
		&comment.Path,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if parentID.Valid {
		comment.ParentID = &parentID.Int64
	}

	return tx.Commit()
}

// buildPath строит путь ltree для нового комментария
func (r *CommentRepository) buildPath(ctx context.Context, tx *sql.Tx, parentID *int64) (string, error) {
	var path string

	if parentID != nil && *parentID > 0 {
		err := tx.QueryRowContext(ctx, "SELECT path FROM comments WHERE id = $1", *parentID).Scan(&path)
		if err != nil {
			return "", fmt.Errorf("parent comment not found: %w", err)
		}
	} else {
		path = ""
	}

	var nextIndex int
	if path == "" {
		query := `SELECT COALESCE(MAX((string_to_array(path::text, '.'))[1]::int), 0) + 1 FROM comments WHERE nlevel(path) = 1`
		if err := tx.QueryRowContext(ctx, query).Scan(&nextIndex); err != nil {
			return "", err
		}
		path = fmt.Sprintf("%d", nextIndex)
	} else {
		// Ищем следующий индекс среди непосредственных детей данного родителя
		query := `SELECT COALESCE(MAX((string_to_array(path::text, '.'))[nlevel($1::ltree) + 1]::int), 0) + 1 FROM comments WHERE $1::ltree @> path AND nlevel(path) = nlevel($1::ltree) + 1`
		if err := tx.QueryRowContext(ctx, query, path).Scan(&nextIndex); err != nil {
			return "", err
		}
		path = fmt.Sprintf("%s.%d", path, nextIndex)
	}

	return path, nil
}

// GetByID получает комментарий по ID
func (r *CommentRepository) GetByID(ctx context.Context, id int64) (*domain.Comment, error) {
	ctx, cancel := context.WithTimeout(ctx, DBQueryTimeout)
	defer cancel()

	query := `
		SELECT id, parent_id, author, content, path, created_at, updated_at
		FROM comments
		WHERE id = $1`

	comment, err := r.scanComment(ctx, query, id)
	if err != nil {
		return nil, err
	}

	return comment, nil
}

// GetRootComments получает корневые комментарии с детьми
func (r *CommentRepository) GetRootComments(ctx context.Context, limit, offset int) ([]domain.Comment, error) {
	ctx, cancel := context.WithTimeout(ctx, DBQueryTimeout)
	defer cancel()

	// Загружаем на 1 больше чтобы проверить есть ли ещё
	query := `
		SELECT id, parent_id, author, content, path, created_at, updated_at
		FROM comments
		WHERE nlevel(path) = 1
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	comments, err := r.scanComments(ctx, query, limit+1, offset)
	if err != nil {
		return nil, err
	}

	// Проверяем есть ли ещё комментарии
	hasMore := len(comments) > limit
	if hasMore {
		comments = comments[:limit]
	}

	// Загружаем всё дерево комментариев одним запросом
	allComments, err := r.loadFullTree(ctx)
	if err != nil {
		return nil, err
	}

	// Строим дерево только для корневых комментариев
	return r.buildTreeForComments(comments, allComments), nil
}

// loadFullTree загружает все комментарии одним запросом
func (r *CommentRepository) loadFullTree(ctx context.Context) (map[int64]*domain.Comment, error) {
	query := `
		SELECT id, parent_id, author, content, path, created_at, updated_at
		FROM comments
		ORDER BY path`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := make(map[int64]*domain.Comment)
	for rows.Next() {
		var comment domain.Comment
		var parentID sql.NullInt64

		err := rows.Scan(
			&comment.ID,
			&parentID,
			&comment.Author,
			&comment.Content,
			&comment.Path,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if parentID.Valid {
			comment.ParentID = &parentID.Int64
		}

		comment.Children = make([]domain.Comment, 0)
		comments[comment.ID] = &comment
	}

	return comments, rows.Err()
}

// buildTreeForComments строит дерево для указанных корневых комментариев
func (r *CommentRepository) buildTreeForComments(rootComments []domain.Comment, allComments map[int64]*domain.Comment) []domain.Comment {
	result := make([]domain.Comment, 0, len(rootComments))

	for _, root := range rootComments {
		// Копируем корневой комментарий
		rootCopy := root
		rootCopy.Children = make([]domain.Comment, 0)

		// Находим всех детей для этого корневого комментария
		for _, comment := range allComments {
			if comment.ParentID != nil && *comment.ParentID == root.ID {
				rootCopy.Children = append(rootCopy.Children, *comment)
			}
		}

		// Рекурсивно заполняем детей
		r.fillChildren(&rootCopy, allComments)
		result = append(result, rootCopy)
	}

	return result
}

// fillChildren рекурсивно заполняет детей
func (r *CommentRepository) fillChildren(comment *domain.Comment, allComments map[int64]*domain.Comment) {
	for i := range comment.Children {
		child := &comment.Children[i]
		for _, gc := range allComments {
			if gc.ParentID != nil && *gc.ParentID == child.ID {
				childCopy := *gc
				childCopy.Children = make([]domain.Comment, 0)
				child.Children = append(child.Children, childCopy)
			}
		}
		// Рекурсивно для внуков
		r.fillChildren(&comment.Children[i], allComments)
	}
}

// GetChildren получает дочерние комментарии
func (r *CommentRepository) GetChildren(ctx context.Context, parentID int64) ([]domain.Comment, error) {
	ctx, cancel := context.WithTimeout(ctx, DBQueryTimeout)
	defer cancel()

	var parentPath string
	err := r.db.QueryRowContext(ctx, "SELECT path FROM comments WHERE id = $1", parentID).Scan(&parentPath)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, parent_id, author, content, path, created_at, updated_at
		FROM comments
		WHERE nlevel(path) = nlevel($1::ltree) + 1 
		  AND subpath(path, 0, nlevel($1::ltree)) = $1::ltree
		ORDER BY path`

	rows, err := r.db.QueryContext(ctx, query, parentPath)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAllComments(rows)
}

// GetTree получает дерево комментариев
func (r *CommentRepository) GetTree(ctx context.Context, parentID *int64, limit, offset int) ([]domain.Comment, error) {
	if parentID == nil {
		return r.GetRootComments(ctx, limit, offset)
	}

	ctx, cancel := context.WithTimeout(ctx, DBQueryTimeout)
	defer cancel()

	var parentPath string
	err := r.db.QueryRowContext(ctx, "SELECT path FROM comments WHERE id = $1", *parentID).Scan(&parentPath)
	if err != nil {
		return nil, domain.ErrCommentNotFound
	}

	// Загружаем все комментарии дерева родителя
	query := `
		SELECT id, parent_id, author, content, path, created_at, updated_at
		FROM comments
		WHERE $1::ltree @> path AND id != $2
		ORDER BY path`

	comments, err := r.scanComments(ctx, query, parentPath, *parentID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Загружаем всё дерево для построения иерархии
	allComments, err := r.loadFullTree(ctx)
	if err != nil {
		return nil, err
	}

	// Строим дерево для загруженных комментариев
	return r.buildTreeForComments(comments, allComments), nil
}

// Delete удаляет комментарий и все вложенные
func (r *CommentRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM comments WHERE id = $1", id)
	return err
}

// Search ищет комментарии по тексту
func (r *CommentRepository) Search(ctx context.Context, query string, limit, offset int) ([]domain.Comment, error) {
	sqlQuery := `
		SELECT id, parent_id, author, content, path, created_at, updated_at
		FROM comments
		WHERE to_tsvector('russian', content) @@ plainto_tsquery('russian', $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	return r.scanComments(ctx, sqlQuery, query, limit, offset)
}

func (r *CommentRepository) scanComment(ctx context.Context, query string, args ...interface{}) (*domain.Comment, error) {
	row := r.db.QueryRowContext(ctx, query, args...)

	var comment domain.Comment
	var parentID sql.NullInt64

	err := row.Scan(
		&comment.ID,
		&parentID,
		&comment.Author,
		&comment.Content,
		&comment.Path,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if parentID.Valid {
		comment.ParentID = &parentID.Int64
	}

	return &comment, nil
}

func (r *CommentRepository) scanComments(ctx context.Context, query string, args ...interface{}) ([]domain.Comment, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAllComments(rows)
}

func (r *CommentRepository) scanAllComments(rows *sql.Rows) ([]domain.Comment, error) {
	var comments []domain.Comment

	for rows.Next() {
		var comment domain.Comment
		var parentID sql.NullInt64

		err := rows.Scan(
			&comment.ID,
			&parentID,
			&comment.Author,
			&comment.Content,
			&comment.Path,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if parentID.Valid {
			comment.ParentID = &parentID.Int64
		}

		comments = append(comments, comment)
	}

	return comments, rows.Err()
}
