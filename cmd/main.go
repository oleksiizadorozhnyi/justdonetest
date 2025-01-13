package main

import (
	"JustDone/internal/config"
	"JustDone/internal/repository/postgres"
	"JustDone/internal/server"
	"JustDone/internal/service"
	"fmt"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
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

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	repo, err := repository.NewOrderRepo(cfg.Postgres)
	if err != nil {
		logger.Fatal(err.Error())
	}

	serviceOrder, err := service.NewOrderService(repo)
	if err != nil {
		logger.Fatal(err.Error())
	}

	srv, err := server.NewServer(serviceOrder)
	if err != nil {
		logger.Fatal(err.Error())
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, os.Interrupt)

	go func() {
		logger.Info(fmt.Sprintf("Server Running on %v...", cfg.ServerPort))
		if err := srv.Run(cfg.ServerPort); err != nil {
			logger.Fatal(err.Error())
		}
	}()

	<-ch
	logger.Info("Received stop signal, shutting down the server...")
}
