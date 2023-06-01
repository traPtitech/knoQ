package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/traPtitech/knoQ/domain"
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
	domain.VERSION = os.Getenv("KNOQ_VERSION")
	domain.REVISION = os.Getenv("KNOQ_REVISION")
	domain.DEVELOPMENT, _ = strconv.ParseBool(os.Getenv("DEVELOPMENT"))

	if os.Getenv("CHANNEL_ID_DAILY") == "" {
		err := os.Setenv("CHANNEL_ID_DAILY", os.Getenv("CHANNEL_ID"))
		if err != nil {
			panic(err)
		}
	}

	if os.Getenv("CHANNEL_ID_ACTIVITY") == "" {
		err := os.Setenv("CHANNEL_ID_ACTIVITY", os.Getenv("CHANNEL_ID"))
		if err != nil {
			panic(err)
		}
	}

	gormRepo := db.GormRepository{}
	err := gormRepo.Setup(os.Getenv("MARIADB_HOSTNAME"), os.Getenv("MARIADB_USERNAME"),
		os.Getenv("MARIADB_PASSWORD"), os.Getenv("MARIADB_DATABASE"), os.Getenv("TOKEN_KEY"), os.Getenv("GORM_LOG_LEVEL"))
	if err != nil {
		panic(err)
	}
	traqRepo := traq.TraQRepository{
		Config: &oauth2.Config{
			ClientID:    os.Getenv("CLIENT_ID"),
			RedirectURL: os.Getenv("ORIGIN") + "/api/callback",
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
		ActivityChannelID: os.Getenv("CHANNEL_ID_ACTIVITY"),
		DailyChannelId:    os.Getenv("CHANNEL_ID_DAILY"),
		Origin:            os.Getenv("ORIGIN"),
	}

	e := handler.SetupRoute()

	// webhook
	job := utils.InitPostEventToTraQ(&repo.GormRepo, handler.WebhookSecret,
		handler.DailyChannelId, handler.WebhookID, handler.Origin)
	_, _ = scheduler.Every().Day().At("08:00").Run(job)

	e.Logger.Info("start")

	// サーバースタート
	go func() {
		if err := e.Start(":3000"); err != nil {
			e.Logger.Info("shutting down the server")
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
