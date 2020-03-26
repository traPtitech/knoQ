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
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/calendar/v3"
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
	googleAPI := &repo.GoogleAPIRepository{
		Config: &jwt.Config{
			Email:      os.Getenv("SERVICE_ACCOUNT_EMAIL"),
			PrivateKey: []byte(os.Getenv("SERVICE_ACCOUNT_KEY")),
			Scopes: []string{
				calendar.CalendarReadonlyScope,
			},
			TokenURL: google.JWTTokenURL,
		},
		CalendarID: os.Getenv("TRAQ_CALENDARID"),
	}
	googleAPI.Setup()

	logger, _ := zap.NewDevelopment()
	handler := &router.Handlers{
		Repo: &repo.GormRepository{
			DB: db,
		},
		InitExternalUserGroupRepo: func(token string, ver repo.TraQVersion) interface {
			repo.UserRepository
			repo.GroupRepository
		} {
			traQRepo := new(repo.TraQRepository)
			traQRepo.Token = token
			traQRepo.Version = ver
			return traQRepo
		},
		ExternalRoomRepo: googleAPI,
		Logger:           logger,
		SessionKey:       SESSION_KEY,
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
