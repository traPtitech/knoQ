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
	"github.com/traPtitech/knoQ/message"
	"github.com/traPtitech/knoQ/repository"

	"github.com/traPtitech/knoQ/utils/tz"
	"golang.org/x/oauth2"

	"github.com/traPtitech/knoQ/router"

	"github.com/gorilla/sessions"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

var (
	version     = getenv("KNOQ_VERSION", "unknown")
	revision    = getenv("KNOQ_REVISION", "unknown")
	development = getenv("DEVELOPMENT", "false")

	mariadbHost     = getenv("MARIADB_HOSTNAME", "db")
	mariadbUser     = getenv("MARIADB_USERNAME", "root")
	mariadbPassword = getenv("MARIADB_PASSWORD", "password")
	mariadbDatabase = getenv("MARIADB_DATABASE", "knoQ")
	mariadbPort     = getenv("MARIADB_PORT", "3306")
	tokenKey        = getenv("TOKEN_KEY", "random32wordsXXXXXXXXXXXXXXXXXXX")
	gormLogLevel    = getenv("GORM_LOG_LEVEL", "silent")

	clientID          = getenv("CLIENT_ID", "client_id")
	origin            = getenv("ORIGIN", "http://localhost:3000")
	sessionKey        = getenv("SESSION_KEY", "random32wordsXXXXXXXXXXXXXXXXXXX")
	webhookID         = getenv("WEBHOOK_ID", "")
	webhookSecret     = getenv("WEBHOOK_SECRET", "")
	activityChannelID = getenv("ACTIVITY_CHANNEL_ID", "")
	dailyChannelID    = getenv("DAILY_CHANNEL_ID", "")

	// TODO: traQにClient Credential Flowが実装されたら定期的に取得するように変更する
	// Issue: https://github.com/traPtitech/traQ/issues/2403
	traqAccessToken = getenv("TRAQ_ACCESS_TOKEN", "")
)

func main() {
	logger, _ := zap.NewDevelopment()
	domain.VERSION = version
	domain.REVISION = revision
	domain.DEVELOPMENT, _ = strconv.ParseBool(development)

	gormRepo := db.GormRepository{}
	err := gormRepo.Setup(mariadbHost, mariadbUser, mariadbPassword, mariadbDatabase, mariadbPort, tokenKey, gormLogLevel, tz.JST)
	if err != nil {
		panic(err)
	}
	traqRepo := traq.TraQRepository{
		Config: &oauth2.Config{
			ClientID:    clientID,
			RedirectURL: origin + "/api/callback",
			Scopes:      []string{"read"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://q.trap.jp/api/v3/oauth2/authorize",
				TokenURL: "https://q.trap.jp/api/v3/oauth2/token",
			},
		},
		URL:               "https://q.trap.jp/api/v3",
		ServerAccessToken: traqAccessToken,
	}
	repo := &repository.Repository{
		GormRepo: gormRepo,
		TraQRepo: traqRepo,
	}
	handler := &router.Handlers{
		Repo:       repo,
		Logger:     logger,
		SessionKey: []byte(sessionKey),
		ClientID:   clientID,
		SessionOption: sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 30,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		},
		WebhookID:         webhookID,
		WebhookSecret:     webhookSecret,
		ActivityChannelID: activityChannelID,
		DailyChannelId:    dailyChannelID,
		Origin:            origin,
	}

	e := handler.SetupRoute()

	// webhook
	c := cron.New(cron.WithLocation(tz.JST))
	_, err = c.AddFunc(
		"0 8 * * *",
		message.InitPostEventToTraQ(
			&repo.GormRepo,
			handler.WebhookSecret,
			handler.DailyChannelId,
			handler.WebhookID,
			handler.Origin,
		),
	)
	if err != nil {
		panic(err)
	}
	c.Start()

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

func getenv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
