package main

import (
	"comment-tree/internal/config"
	"comment-tree/internal/handler"
	"comment-tree/internal/infrastructure/postgres"
	"comment-tree/internal/router"
	"comment-tree/internal/usecase"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/logger"
)

func main() {
	// 1. Загрузка конфигурации
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Инициализация логгера
	logInstance, err := logger.InitLogger(
		logger.ZapEngine,
		"comment-tree",
		"development",
		logger.WithLevel(logger.InfoLevel),
	)
	if err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}
	logInstance.Info("Starting Comment Tree service")

	// 3. Подключение к PostgreSQL
	dbPool, err := dbpg.New(cfg.Database.DSN, nil, &dbpg.Options{
		MaxOpenConns: 10,
		MaxIdleConns: 5,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Проверка подключения к БД
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := dbPool.Master.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	logInstance.Info("Connected to PostgreSQL")

	// 4. Создание репозитория
	commentRepo := postgres.NewCommentRepository(dbPool.Master)

	// 5. Создание use-case
	commentUC := usecase.NewCommentUseCase(commentRepo)

	// 6. Создание хендлера
	h := handler.NewHandler(commentUC)

	// 7. Создание роутера
	r := router.NewRouter(h)

	// 8. Запуск сервера
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	logInstance.Info("Server starting", "address", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
