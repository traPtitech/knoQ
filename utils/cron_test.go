package utils

import (
	repo "room/repository"
	"testing"
	"time"

	"github.com/carlescere/scheduler"
)

func Test_initPostEventToTraQ(t *testing.T) {
	db, _ := repo.SetupDatabase()
	gr := &repo.GormRepository{
		DB: db,
	}

	job := initPostEventToTraQ(gr)

	scheduler.Every(2).Seconds().Run(job)
	time.Sleep(time.Second * 10)
}
