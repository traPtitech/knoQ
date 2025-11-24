package db

import (
	"errors"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"
	"github.com/traPtitech/knoQ/domain"
	"gorm.io/gorm"
)

func userPreload(tx *gorm.DB) *gorm.DB {
	return tx.Preload("Provider")
}

// LoginUser のときのみ，プロバイダ，トークン情報が含まれる
func (repo *gormRepository) SaveUser(args domain.SaveUserArgs) (*domain.User, error) {
	user := User{
		ID:    args.UserID,
		State: args.State,
		Token: Token{
			UserID: args.UserID,
			Oauth2Token: &Oauth2Token{
				AccessToken:  args.AccessToken,
				TokenType:    args.TokenType,
				RefreshToken: args.RefreshToken,
				Expiry:       args.Expiry,
			},
		},
		Provider: Provider{
			UserID:  args.UserID,
			Issuer:  args.Issuer,
			Subject: args.Subject,
		},
	}
	u, err := saveUser(repo.db, &user)
	du := convUserTodomainUser(*u)
	return &du, defaultErrorHandling(err)
}

func (repo *gormRepository) UpdateiCalSecret(userID uuid.UUID, secret string) error {
	err := updateiCalSecret(repo.db, userID, secret)
	return defaultErrorHandling(err)
}

func (repo *gormRepository) GetUser(userID uuid.UUID) (*domain.User, error) {
	u, err := getUser(userPreload(repo.db), userID)
	du := convUserTodomainUser(*u)
	return &du, defaultErrorHandling(err)
}

func (repo *gormRepository) GetAllUsers(onlyActive bool) ([]*domain.User, error) {
	us, err := getAllUsers(userPreload(repo.db), onlyActive)
	dus := lo.Map(us, func(u *User, _ int) *domain.User {
		return &domain.User{
			ID:         u.ID,
			Privileged: u.Privilege,
			State:      u.State,
		}
	})
	return dus, defaultErrorHandling(err)
}

func (repo *gormRepository) SyncUsers(argss []domain.SyncUserArgs) error {

	err := repo.db.Transaction(func(tx *gorm.DB) error {
		existingUsers, err := getAllUsers(tx, false)
		if err != nil {
			return err
		}

		for _, args := range argss {
			exist := false
			for _, eu := range existingUsers {
				if args.UserID == eu.ID {
					// contains
					exist = true
					if args.State != eu.State {
						eu.State = args.State
						_, err := saveUser(tx, eu)
						if err != nil {
							return err
						}
					}
					break
				}
			}
			u := &User{
				ID:    args.UserID,
				State: args.State,
				Provider: Provider{
					UserID:  args.UserID,
					Issuer:  args.Issuer,
					Subject: args.Subject,
				},
			}
			if !exist {
				_, err := saveUser(tx, u)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})

	return defaultErrorHandling(err)
}

// saveUser user.IcalSecret == "" の時、値は更新されません。
// また、user.Provider, user.Tokenは空の時、更新されません。
// user.Privilegeは常に更新されません。
func saveUser(db *gorm.DB, user *User) (*User, error) {
	err := db.Transaction(func(tx *gorm.DB) error {
		existingUser, err := getUser(tx.Preload("Provider").Preload("Token"), user.ID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&user).Error
		}
		if err != nil {
			return err
		}

		if user.IcalSecret == "" {
			user.IcalSecret = existingUser.IcalSecret
		}
		return tx.Session(&gorm.Session{FullSaveAssociations: true}).Updates(user).Error
	})

	return user, err
}

func updateiCalSecret(db *gorm.DB, userID uuid.UUID, secret string) error {
	return db.Model(&User{ID: userID}).Update("ical_secret", secret).Error
}

func getUser(db *gorm.DB, userID uuid.UUID) (*User, error) {
	user := User{}
	err := db.Take(&user, userID).Error
	return &user, err
}

func getAllUsers(db *gorm.DB, onlyActive bool) ([]*User, error) {
	users := make([]*User, 0)
	if onlyActive {
		err := db.Where("state = ?", 1).Find(&users).Error
		return users, err
	}
	err := db.Find(&users).Error
	return users, err
}

func (repo *gormRepository) GrantPrivilege(userID uuid.UUID) error {
	err := grantPrivilege(repo.db, userID)
	return defaultErrorHandling(err)
}

func grantPrivilege(db *gorm.DB, userID uuid.UUID) error {
	err := db.Model(&User{ID: userID}).Update("privilege", true).Error
	return err
}

func (repo *gormRepository) GetICalSecret(userID uuid.UUID) (string, error) {
	u, err := getUser(repo.db, userID)
	if err != nil {
		return "", defaultErrorHandling(err)
	}

	return u.IcalSecret, nil
}
