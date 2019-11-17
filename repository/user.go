package repository

// GetUser ユーザー情報を取得します
func GetUser(id string) (*User, error) {
	user := User{}

	// DBに登録されていない場合(初めてアクセスした場合)はDBにレコードを作成する
	if err := DB.FirstOrCreate(&user, &User{TRAQID: id}).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// changeUserToAdmin ユーザーの管理者権限の有無を変更します
func changeUserToAdmin(id string, isAdmin bool) error {
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
