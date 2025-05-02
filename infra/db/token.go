package db

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"

	"github.com/gofrs/uuid"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

func (repo *GormRepository) GetToken(userID uuid.UUID) (*oauth2.Token, error) {
	return getToken(repo.db, userID)
}

// TODO: string を返した方が良い?
func getToken(db *gorm.DB, userID uuid.UUID) (*oauth2.Token, error) {
	u := new(User)
	err := db.Take(u, userID).Error
	if err != nil {
		return nil, defaultErrorHandling(err)
	}

	// decrypt
	if u.AccessToken != "" {
		u.AccessToken, err = decryptByGCM(tokenKey, []byte(u.AccessToken))
		if err != nil {
			return nil, defaultErrorHandling(err)
		}
	}
	return &oauth2.Token{
		AccessToken: u.AccessToken,
	}, defaultErrorHandling(err)
}

// GCM encryption
func encryptByGCM(key []byte, plainText string) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize()) // Unique nonce is required(NonceSize 12byte)
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, err
	}

	cipherText := gcm.Seal(nil, nonce, []byte(plainText), nil)
	cipherText = append(nonce, cipherText...)

	return cipherText, nil
}

// Decrypt by GCM
func decryptByGCM(key []byte, cipherText []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := cipherText[:gcm.NonceSize()]
	plainByte, err := gcm.Open(nil, nonce, cipherText[gcm.NonceSize():], nil)
	if err != nil {
		return "", err
	}

	return string(plainByte), nil
}
