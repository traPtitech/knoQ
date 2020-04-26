package repository

import (
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid"
	jsoniter "github.com/json-iterator/go"
	traQrouterV1 "github.com/traPtitech/traQ/router/v1"
	traQrouterV3 "github.com/traPtitech/traQ/router/v3"
	traQutils "github.com/traPtitech/traQ/utils"
)

var traQjson = jsoniter.Config{
	EscapeHTML:             true,
	SortMapKeys:            true,
	ValidateJsonRawMessage: true,
	TagKey:                 "traq",
}.Froze()

type UserRepository interface {
	CreateUser(isAdmin bool) (*User, error)
	UpdateiCalSecretUser(userID uuid.UUID, secret string) error
	GetUser(userID uuid.UUID) (*User, error)
	GetAllUsers() ([]*User, error)
}

// GormRepository implements UserRepository

func (repo *GormRepository) CreateUser(isAdmin bool) (*User, error) {
	userID, _ := uuid.NewV4()
	user := User{
		ID:    userID,
		Admin: isAdmin,
	}
	if err := repo.DB.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUser ユーザー情報を取得します(なければ作成)
func (repo *GormRepository) GetUser(userID uuid.UUID) (*User, error) {
	user := new(User)
	if err := repo.DB.FirstOrCreate(&user, &User{ID: userID}).Error; err != nil {
		return nil, err
	}
	return user, nil

}

func (repo *GormRepository) GetAllUsers() ([]*User, error) {
	users := make([]*User, 0)
	err := repo.DB.Find(&users).Error
	return users, err
}

func (repo *GormRepository) UpdateiCalSecretUser(userID uuid.UUID, secret string) error {
	if userID == uuid.Nil {
		return ErrNilID
	}
	if err := repo.DB.Model(&User{ID: userID}).Update("ical_secret", secret).Error; err != nil {
		return err
	}
	return nil
}

// traQRepository implements UserRepository

// CreateUser does not use isAdmin and
// create random name and password
func (repo *TraQRepository) CreateUser(isAdmin bool) (*User, error) {
	if repo.Version != TraQv1 {
		repo.Version = TraQv1
		defer func() {
			repo.Version = TraQv3
		}()
	}
	reqUser := &traQrouterV1.PostUserRequest{
		Name:     traQutils.RandAlphabetAndNumberString(20),
		Password: traQutils.RandAlphabetAndNumberString(20),
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
	return &User{ID: traQuser.ID}, nil
}

// GetUser get from /users/{userID}
func (repo *TraQRepository) GetUser(userID uuid.UUID) (*User, error) {
	data, err := repo.getRequest(fmt.Sprintf("/users/%s", userID))
	if err != nil {
		return nil, err
	}
	traQuser := new(traQrouterV3.User)
	err = json.Unmarshal(data, &traQuser)
	return formatV3User(traQuser), err
}

// GetAllUsers get from /users
func (repo *TraQRepository) GetAllUsers() ([]*User, error) {
	data, err := repo.getRequest("/users")
	if err != nil {
		return nil, err
	}
	traQusers := make([]*traQrouterV3.User, 0)
	err = traQjson.Unmarshal(data, &traQusers)
	users := make([]*User, len(traQusers))
	for i, u := range traQusers {
		users[i] = formatV3User(u)
	}
	return users, err
}
func (repo *TraQRepository) UpdateiCalSecretUser(userID uuid.UUID, secret string) error {
	return ErrForbidden
}

func formatV3User(u *traQrouterV3.User) *User {
	return &User{
		ID:          u.ID,
		Admin:       false,
		Name:        u.Name,
		DisplayName: u.DisplayName,
	}
}

// GetUser ユーザー情報を取得します(なければ作成)
func GetUser(id uuid.UUID) (User, error) {
	user := User{}

	// DBに登録されていない場合(初めてアクセスした場合)はDBにレコードを作成する
	if err := DB.FirstOrCreate(&user, &User{ID: id}).Error; err != nil {
		return User{}, err
	}
	return user, nil
}

// changeUserToAdmin ユーザーの管理者権限の有無を変更します
func changeUserToAdmin(id uuid.UUID, isAdmin bool) error {
	// ユーザー取得
	user, err := GetUser(id)
	if err != nil {
		return err
	}

	// 変更
	if err := DB.Model(user).Update("admin", isAdmin).Error; err != nil {
		return err
	}
	return nil
}
