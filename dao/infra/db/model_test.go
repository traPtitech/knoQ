package db

import (
	"testing"
)

func TestGormRepository_Setup(t *testing.T) {
	gormRepo := new(GormRepository)
	err := gormRepo.Setup()
	if err != nil {
		panic(err)
	}
}
