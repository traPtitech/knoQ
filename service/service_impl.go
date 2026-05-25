package service

import (
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/infra/traq"
)

type service struct {
	GormRepo  domain.Repository
	TraQRepo  *traq.TraQRepository
	TxManager domain.TransactionManager
}

// implements domain

func NewService(repo domain.Repository, traqRepo *traq.TraQRepository, txManager domain.TransactionManager) domain.Service {
	return &service{GormRepo: repo, TraQRepo: traqRepo, TxManager: txManager}
}
