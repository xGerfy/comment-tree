package router

import (
	"comment-tree/internal/handler"
	"comment-tree/internal/middleware"
	"time"

	"github.com/wb-go/wbf/ginext"
)

// NewRouter создает новый роутер со всеми middleware и маршрутами
func NewRouter(h *handler.Handler) *ginext.Engine {
	router := ginext.New("release")
	router.Use(ginext.Recovery())

	// Rate limiting: 10 запросов в секунду для API
	rateLimiter := middleware.NewRateLimiter(10, time.Second)

	// API группа с rate limiting
	api := router.Group("/api")
	api.Use(rateLimiter.Middleware())
	{
		api.POST("/comments", h.CreateComment)
		api.GET("/comments", h.GetComments)
		api.GET("/comments/:id", h.GetComment)
		api.DELETE("/comments/:id", h.DeleteComment)
		api.POST("/comments/search", h.SearchComments)
	}

	// Веб-интерфейс
	router.StaticFile("/", "web/index.html")
	router.Static("/web", "web")

	// UI для всех остальных роутов
	router.NoRoute(func(c *ginext.Context) {
		c.File("web/index.html")
	})

	return router
}
