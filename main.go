package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/traPtitech/knoQ/infra/db"
	"github.com/traPtitech/knoQ/infra/traq"
	"github.com/traPtitech/knoQ/usecase/production"
	"github.com/traPtitech/knoQ/utils"
	"golang.org/x/oauth2"

	"github.com/traPtitech/knoQ/router"

	"github.com/carlescere/scheduler"
	"github.com/gorilla/sessions"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment()
	gormRepo := db.GormRepository{}
	err := gormRepo.Setup(os.Getenv("MARIADB_HOSTNAME"), os.Getenv("MARIADB_USERNAME"),
		os.Getenv("MARIADB_PASSWORD"), os.Getenv("MARIADB_DATABASE"))
	if err != nil {
		panic(err)
	}
	traqRepo := traq.TraQRepository{
		Config: &oauth2.Config{
			ClientID:    os.Getenv("CLIENT_ID"),
			RedirectURL: "http://localhost:6006/api/callback",
			Scopes:      []string{"read"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://q.trap.jp/api/v3/oauth2/authorize",
				TokenURL: "https://q.trap.jp/api/v3/oauth2/token",
			},
		},
		URL: "https://q.trap.jp/api/v3",
	}
	repo := &production.Repository{
		GormRepo: gormRepo,
		TraQRepo: traqRepo,
	}
	handler := &router.Handlers{
		Repo:       repo,
		Logger:     logger,
		SessionKey: []byte(os.Getenv("SESSION_KEY")),
		ClientID:   os.Getenv("CLIENT_ID"),
		SessionOption: sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 30,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		},
		WebhookID:         os.Getenv("WEBHOOK_ID"),
		WebhookSecret:     os.Getenv("WEBHOOK_SECRET"),
		ActivityChannelID: os.Getenv("CHANNEL_ID"),
		Origin:            os.Getenv("ORIGIN"),
	}

	e := handler.SetupRoute()

	// webhook
	job := utils.InitPostEventToTraQ(&repo.GormRepo, handler.WebhookSecret,
		handler.ActivityChannelID, handler.WebhookID, handler.Origin)
	scheduler.Every().Day().At("08:00").Run(job)

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
