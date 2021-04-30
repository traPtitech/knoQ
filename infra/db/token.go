package db

import (
	"github.com/gofrs/uuid"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

func (repo *GormRepository) GetToken(userID uuid.UUID) (*oauth2.Token, error) {
	return getToken(repo.db, userID)
}

func getToken(db *gorm.DB, userID uuid.UUID) (*oauth2.Token, error) {
	token := Token{}
	err := db.Take(&token, userID).Error
	// decrypt
	return token.Token, defaultErrorHandling(err)
}
