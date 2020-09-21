package service

import (
	repo "room/repository"

	"github.com/gofrs/uuid"
)

// User is max User struct
type User struct {
	ID          uuid.UUID
	Name        string
	DisplayName string
	Admin       bool
}

func formatUser(meta *repo.UserMeta, body *repo.UserBody) *User {
	return &User{
		ID:          meta.ID,
		Name:        body.Name,
		DisplayName: body.DisplayName,
		Admin:       meta.Admin,
	}

}

// formatUsers metaにbodyを結び付ける
func formatUsers(meta []*repo.UserMeta, body []*repo.UserBody) []*User {
	bodyMap := map[uuid.UUID]*repo.UserBody{}
	for _, b := range body {
		bodyMap[b.ID] = b
	}
	users := make([]*User, 0, len(meta))
	for _, m := range meta {
		// TODO consider state=0 users
		b := bodyMap[m.ID]
		if b != nil {
			users = append(users, formatUser(m, b))
		}
	}
	return users
}

func (d Dao) GetUser(token string, userID uuid.UUID) (*User, error) {
	UserGroupRepo := d.InitExternalUserGroupRepo(token, repo.TraQv3)

	userBody, err := UserGroupRepo.GetUser(userID)
	if err != nil {
		//if err.Error() == http.StatusText(http.StatusUnauthorized) {
		//// 認証が切れている
		//return nil, repo.ErrForbidden
		//}
		return nil, err
	}
	userMeta, _ := d.Repo.GetUser(userID)
	user := formatUser(userMeta, userBody)
	return user, nil
}

func (d Dao) GetAllUsers(token string) ([]*User, error) {
	UserGroupRepo := d.InitExternalUserGroupRepo(token, repo.TraQv3)

	body, err := UserGroupRepo.GetAllUsers()
	if err != nil {
		//if err.Error() == http.StatusText(http.StatusUnauthorized) {
		//// 認証が切れている
		//return nil, repo.ErrForbidden
		//}
		return nil, err
	}
	meta, err := d.Repo.GetAllUsers()

	return formatUsers(meta, body), err
}
