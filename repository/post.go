package repository

import (
	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/infra/db"
)

func (repo *Repository) CreatePost(params domain.WritePostParams) (*domain.Post, error) {
	p := db.WritePostParams{
		WritePostParams: params,
	}
	post, err := repo.GormRepo.CreatePost(p)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	domainPost := db.ConvPostTodomainPost(*post)
	return &domainPost, nil
}

func (repo *Repository) GetPost(MessageID uuid.UUID) (*domain.Post, error) {
	post, err := repo.GormRepo.GetPost(MessageID)
	if err != nil {
		return nil, defaultErrorHandling(err)
	}
	domainPost := db.ConvPostTodomainPost(*post)
	return &domainPost, nil
}
