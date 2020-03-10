package repository

import "github.com/gofrs/uuid"

type UserRepository interface {
	CreateUser(userID uuid.UUID, isAdmin bool) (*User, error)
	GetUser(userID uuid.UUID) (*User, error)
	GetAllUsers() ([]*User, error)
}

func (repo *GormRepository) CreateUser(userID uuid.UUID, isAdmin bool) (*User, error) {
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
