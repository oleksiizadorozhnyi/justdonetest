package main

import (
	repository "JustDone/internal/repository/postgres"
	"JustDone/internal/server"
	service2 "JustDone/internal/service"
	"fmt"
	"go.uber.org/zap"
	"log"
	"os"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			zap.L().Error("RECOVERED FROM PANIC", zap.Any("error", r))
			fmt.Println("RECOVERED FROM PANIC", r)
		}
	}()

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)

	dbString := os.Getenv("DATABASE_URL")
	if dbString == "" {
		log.Fatalf("DATABASE_URL is not set")
	}

	repo, err := repository.NewOrderRepo(dbString)
	if err != nil {
		logger.Fatal(err.Error())
	}

	service, err := service2.NewOrderService(repo)
	if err != nil {
		logger.Fatal(err.Error())
	}

	srv, err := server.NewServer(service)
	if err != nil {
		logger.Fatal(err.Error())
	}

	fmt.Println("Server Running on 8080...")
	if err := srv.Run(":8080"); err != nil {
		logger.Fatal(err.Error())
	}
}
