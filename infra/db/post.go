package db

import (
	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"gorm.io/gorm"
)

type WritePostParams struct {
	domain.WritePostParams
}

func (repo *GormRepository) CreatePost(params WritePostParams) (*Post, error) {
	p, err := createPost(repo.db, params)
	return p, defaultErrorHandling(err)
}

func (repo *GormRepository) GetPost(MessageID uuid.UUID) (*Post, error) {
	post, err := getPost(repo.db, MessageID)
	return post, defaultErrorHandling(err)
}

func createPost(db *gorm.DB, params WritePostParams) (*Post, error) {
	post := ConvWritePostParamsToPost(params)
	err := db.Create(&post).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func getPost(db *gorm.DB, messageID uuid.UUID) (*Post, error) {
	post := Post{}
	err := db.Take(&post, messageID).Error
	return &post, err
}
