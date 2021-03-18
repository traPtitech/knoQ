package production

import (
	"github.com/traPtitech/knoQ/domain"
	"github.com/traPtitech/knoQ/infra/db"
)

func (repo *Repository) SyncUser(params []domain.WriteUserParams, info *domain.ConInfo) ([]*domain.User, error) {
	// 1. 特権による同期
	// 1.1 特権か？
	// 1.2 変換
	// 1.3 いないユーザーを作っていく(UserMeta, UserBody)

	return nil, nil
}

func (repo *Repository) LoginUser(query, state, codeVerifier string) (*domain.User, error) {
	// 2. ログインによる作成
	// 2.1 traQ からOAuthの情報を使ってユーザーを識別
	// 2.1 ユーザーが存在しなければ、作成。存在すれば、トークンを更新
	//     Provider, Tokenにも適切な値を入れる

	t, err := repo.traQRepo.GetOAuthToken(query, state, codeVerifier)
	if err != nil {
		return nil, err
	}
	traQUser, err := repo.traQRepo.GetUserMe(t)
	if err != nil {
		return nil, err
	}
	userMeta := &db.UserMeta{
		ID: traQUser.ID,
		Token: db.Token{
			UserID: traQUser.ID,
			Token:  t,
		},
		Provider: db.Provider{
			UserID:  traQUser.ID,
			Issuer:  "traQ",
			Subject: traQUser.ID.String(),
		},
	}
	userMeta, err = repo.gormRepo.SaveUser(*userMeta)
	if err != nil {
		return nil, err
	}

	user := ConvertPointerv3UserToPointerdomainUser(traQUser)
	user.Icon = repo.traQRepo.URL + "/public/icon/" + user.Name
	user.Privileged = userMeta.Privilege
	user.IsTrap = true

	return user, nil
}
