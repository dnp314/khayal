package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dnp314/khayal/internal/controller"
	"github.com/dnp314/khayal/internal/repository"
	"github.com/dnp314/khayal/internal/service"
	"github.com/jackc/pgx/v5"
)

type Controllers struct {
	Questions *controller.QuestionController
	Khayals   *controller.KhayalController
	Auth      *controller.AuthController
}

func main() {

	urlExample := "postgres://khayal:khayal@localhost:5432/khayal"
	conn, err := pgx.Connect(context.Background(), urlExample)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

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

	router := createRouter(controllers)

	server := http.Server{
		Addr:    ":8000",
		Handler: router,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal("server did not start")
	}
}
