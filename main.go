package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	repo "room/repository"
	"room/router"
	"time"

	"go.uber.org/zap"
)

var (
	SESSION_KEY = []byte(os.Getenv("SESSION_KEY"))
)

func main() {
	db, err := repo.SetupDatabase()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	logger, _ := zap.NewDevelopment()
	handler := &router.Handlers{
		ExternalUserRepo: &repo.TraQRepository{
			APIRepository: repo.APIRepository{
				BaseURL: "https://q.trap.jp/api/1.0",
			},
		},
		Logger:     logger,
		SessionKey: SESSION_KEY,
	}

	e := handler.SetupRoute(db)

	// サーバースタート
	go func() {
		if err := e.Start(":3000"); err != nil {
			e.Logger.Info("shutting down the server")
		}
	}()
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
