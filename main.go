package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	repo "room/repository"
	"room/router"
	"room/router/service"
	"time"

	"github.com/gorilla/sessions"
	"go.uber.org/zap"
	"golang.org/x/oauth2/google"
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
		CalendarID: os.Getenv("TRAQ_CALENDARID"),
	}
	bytes, err := ioutil.ReadFile("service.json")
	if err != nil {
		panic("service.json does not exist.")
	}
	googleAPI.Config, err = google.JWTConfigFromJSON(bytes, calendar.CalendarReadonlyScope)
	if err != nil {
		panic(err)
	}

	googleAPI.Setup()

	logger, _ := zap.NewDevelopment()
	handler := &router.Handlers{
		Dao: service.Dao{
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
				traQRepo.Host = "https://q.trap.jp/api"
				traQRepo.NewRequest = traQRepo.DefaultNewRequest
				return traQRepo
			},
			InitTraPGroupRepo: func(token string, ver repo.TraQVersion) interface {
				repo.GroupRepository
			} {
				traPGroupRepo := new(repo.TraPGroupRepository)
				traPGroupRepo.Token = token
				traPGroupRepo.Version = ver
				traPGroupRepo.Host = "https://q.trap.jp/api"
				traPGroupRepo.NewRequest = traPGroupRepo.DefaultNewRequest
				return traPGroupRepo
			},
			ExternalRoomRepo: googleAPI,
		},
		Logger:     logger,
		SessionKey: SESSION_KEY,
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
		Origin:            os.Getenv("ROOM_ORIGIN"),
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
