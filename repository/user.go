package repository

import (
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid"
	jsoniter "github.com/json-iterator/go"
	traQrouterV1 "github.com/traPtitech/traQ/router/v1"
	traQrouterV3 "github.com/traPtitech/traQ/router/v3"
)

var traQjson = jsoniter.Config{
	EscapeHTML:             true,
	SortMapKeys:            true,
	ValidateJsonRawMessage: true,
	TagKey:                 "traq",
}.Froze()

type UserMetaRepository interface {
	SaveUser(isAdmin bool) (*UserMeta, error)
	GetUser(userID uuid.UUID) (*UserMeta, error)
	GetAllUsers() ([]*UserMeta, error)
	ReplaceToken(userID uuid.UUID, token string) error
	GetToken(userID uuid.UUID) (string, error)
	UpdateiCalSecretUser(userID uuid.UUID, secret string) error
}

type UserBodyRepository interface {
	CreateUser(name, displayName, password string) (*UserBody, error)
	GetUser(userID uuid.UUID) (*UserBody, error)
	GetAllUsers() ([]*UserBody, error)
}

// GormRepository implements UserRepository

func (repo *GormRepository) SaveUser(isAdmin bool) (*UserMeta, error) {
	userID, _ := uuid.NewV4()
	user := UserMeta{
		ID:    userID,
		Admin: isAdmin,
	}
	if err := repo.DB.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUser ユーザー情報を取得します(なければ作成)
func (repo *GormRepository) GetUser(userID uuid.UUID) (*UserMeta, error) {
	user := new(UserMeta)
	if err := repo.DB.FirstOrCreate(&user, &UserMeta{ID: userID}).Error; err != nil {
		return nil, err
	}
	return user, nil

}

func (repo *GormRepository) GetAllUsers() ([]*UserMeta, error) {
	users := make([]*UserMeta, 0)
	err := repo.DB.Find(&users).Error
	return users, err
}

func (repo *GormRepository) UpdateiCalSecretUser(userID uuid.UUID, secret string) error {
	if userID == uuid.Nil {
		return ErrNilID
	}
	if err := repo.DB.Model(&UserMeta{ID: userID}).Update("ical_secret", secret).Error; err != nil {
		return err
	}
	return nil
}

func (repo *GormRepository) ReplaceToken(userID uuid.UUID, token string) error {
	user := UserMeta{
		ID: userID,
	}
	return repo.DB.Model(&user).Update("token", token).Error
}

func (repo *GormRepository) GetToken(userID uuid.UUID) (string, error) {
	user := UserMeta{
		ID: userID,
	}
	err := repo.DB.First(&user).Error
	if err != nil {
		return "", err
	}

	return user.Token, nil
}

// traQRepository implements UserRepository

// CreateUser 新たにユーザーを作成する
func (repo *TraQRepository) CreateUser(name, password, displayName string) (*UserBody, error) {
	if repo.Version != TraQv1 {
		repo.Version = TraQv1
		defer func() {
			repo.Version = TraQv3
		}()
	}
	reqUser := &traQrouterV1.PostUserRequest{
		Name:     name,
		Password: password,
	}
	body, _ := json.Marshal(reqUser)
	resBody, err := repo.postRequest("/users", body)
	if err != nil {
		return nil, err
	}
	traQuser := struct {
		ID uuid.UUID `json:"id"`
	}{}
	err = json.Unmarshal(resBody, &traQuser)
	if err != nil {
		return nil, err
	}
	return &UserBody{ID: traQuser.ID}, nil
}

// GetUser get from /users/{userID}
func (repo *TraQRepository) GetUser(userID uuid.UUID) (*UserBody, error) {
	data, err := repo.getRequest(fmt.Sprintf("/users/%s", userID))
	if err != nil {
		return nil, err
	}
	traQuser := new(traQrouterV3.User)
	err = json.Unmarshal(data, &traQuser)
	return formatV3User(traQuser), err
}

// GetAllUsers get from /users
func (repo *TraQRepository) GetAllUsers() ([]*UserBody, error) {
	data, err := repo.getRequest("/users")
	if err != nil {
		return nil, err
	}
	traQusers := make([]*traQrouterV3.User, 0)
	err = traQjson.Unmarshal(data, &traQusers)
	users := make([]*UserBody, len(traQusers))
	for i, u := range traQusers {
		users[i] = formatV3User(u)
	}
	return users, err
}
func (repo *TraQRepository) UpdateiCalSecretUser(userID uuid.UUID, secret string) error {
	return ErrForbidden
}

func formatV3User(u *traQrouterV3.User) *UserBody {
	return &UserBody{
		ID:          u.ID,
		Name:        u.Name,
		DisplayName: u.DisplayName,
	}
}
