package db

import (
	"github.com/gofrs/uuid"
	"golang.org/x/oauth2"
)

func (repo *GormRepository) GetToken(userID uuid.UUID) (*oauth2.Token, error) {
	token := Token{
		UserID: userID,
	}
	err := repo.db.Take(&token).Error
	return token.Token, err
}
