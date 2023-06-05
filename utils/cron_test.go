package utils

import (
	"github.com/traPtitech/knoQ/infra/db"

	"os"
	"testing"
)

func Test_initPostEventToTraQ(t *testing.T) {
	gormRepo := db.GormRepository{}
	err := gormRepo.Setup(os.Getenv("MARIADB_HOSTNAME"), os.Getenv("MARIADB_USERNAME"),
		os.Getenv("MARIADB_PASSWORD"), os.Getenv("MARIADB_DATABASE"), os.Getenv("TOKEN_KEY"), "silent")
	if err != nil {
		panic(err)
	}

	InitPostEventToTraQ(&gormRepo, os.Getenv("WEBHOOK_SECRET"),
		os.Getenv("ACTIVITY_CHANNEL_ID"), os.Getenv("WEBHOOK_ID"), os.Getenv("ORIGIN"))()
}
