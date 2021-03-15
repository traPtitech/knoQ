package production

import "github.com/traPtitech/knoQ/domain"

func (repo *Repository) SyncUser(params []domain.WriteUserParams, info *domain.ConInfo) ([]*domain.User, error) {
	// 1. 特権による同期
	// 1.1 特権か？
	// 1.2 変換
	// 1.3 いないユーザーを作っていく(UserMeta, UserBody)

	return nil, nil
}

func (repo *Repository) LoginUser(code, codeVerifier string) (*domain.User, error) {
	// 2. ログインによる作成
	// 2.1 traQ からOAuthの情報を使ってユーザーを識別
	// 2.1 ユーザーが存在しなければ、作成。存在すれば、トークンを更新
	//     Provider, Tokenにも適切な値を入れる

	return nil, nil
}
