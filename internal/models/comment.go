package models

// CreateCommentRequest — запрос на создание комментария
type CreateCommentRequest struct {
	ParentID *int64 `json:"parent_id,omitempty"`
	Author   string `json:"author"`
	Content  string `json:"content"`
}

// CreateCommentResponse — ответ на создание комментария
type CreateCommentResponse struct {
	ID      int64  `json:"id"`
	Author  string `json:"author"`
	Content string `json:"content"`
}

// SearchRequest — запрос на поиск
type SearchRequest struct {
	Query  string `json:"query"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

// CommentsResponse — ответ со списком комментариев
type CommentsResponse struct {
	Comments []interface{} `json:"comments"`
}
