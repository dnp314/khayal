package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/dnp314/khayal/internal/controller"
	"github.com/dnp314/khayal/internal/middleware"
	"github.com/dnp314/khayal/internal/repository"
	"github.com/dnp314/khayal/internal/service"
	"github.com/jackc/pgx/v5"
)

type Controllers struct {
	Questions *controller.QuestionController
	Khayals   *controller.KhayalController
	Auth      *controller.AuthController
}

const port = "8000"

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	ctx := context.Background()
	urlExample := "postgres://khayal:khayal@localhost:5432/khayal"
	conn, err := pgx.Connect(ctx, urlExample)
	if err != nil {
		slog.Error("unable to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := conn.Close(ctx); err != nil {
			slog.Error("unable to close database connection", "error", err)
		}
	}()

	// go utils.StartReturnTimeChecker(conn)

	questionRepo := repository.NewQuestionRepostiory(conn)
	khayalRepo := repository.NewKhayalRepostiory(conn)
	authRepo := repository.NewAuthRepostiory(conn)

	questionService := service.NewQuestionService(questionRepo)
	khayalService := service.NewKhayalService(khayalRepo, questionRepo)
	authService := service.NewAuthService(authRepo)

	khayalController := controller.NewKhayalController(khayalService, questionService)
	questionController := controller.NewQuestionController(questionService, khayalService)
	authController := controller.NewAuthController(authService)

	controllers := Controllers{
		Questions: questionController,
		Khayals:   khayalController,
		Auth:      authController,
	}

	router := createRouter(controllers, middleware.Auth(authService))

	server := http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	slog.Info("starting server", "addr", server.Addr)
	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}
