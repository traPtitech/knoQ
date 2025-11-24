package service

import (
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/infra/traq"
)

type service struct {
	GormRepo domain.Repository
	TraQRepo *traq.TraQRepository
}

// implements domain

func NewService(repo domain.Repository, traqRepo *traq.TraQRepository) domain.Service {
	return &service{GormRepo: repo, TraQRepo: traqRepo}
}
