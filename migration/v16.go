package migration

import (
	"database/sql"
	// gormigrateに必要なその他のimportがあれば追加
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
	// 他の必要なimport (例: time, etc. Model structが含むフィールドによる)
)

// User モデルの定義（マイグレーションに必要な最小限のフィールドのみ）
// アプリケーション内の実際のUserモデルを正確に反映させてください
type User struct {
	ID uuid.UUID `gorm:"type:char(36);primaryKey"`
	// User モデルにある他のフィールド（例: CreatedAt, UpdatedAtなど）
	// CreatedAt time.Time
	// UpdatedAt time.Time
}

// v16Group はマイグレーション後のGroupモデルを反映します
// gormigrateでは、そのマイグレーション時点でのモデルの状態を定義します
type v16Group struct {
	ID          uuid.UUID     `gorm:"type:char(36);primaryKey"`
	Name        string        `gorm:"type:varchar(32);not null"`
	Description string        `gorm:"type:TEXT"`
	IsTraqGroup bool          `gorm:"not null"`
	JoinFreely  sql.NullBool  `gorm:""`
	TraqID      uuid.NullUUID `gorm:""`
	// GroupMemberとGroupAdmin構造体からmany2manyへの変更
	Members        []*User       `gorm:"many2many:group_members;"` // 結合テーブル名を明示
	Admins         []*User       `gorm:"many2many:group_admins;"`  // 結合テーブル名を明示
	CreatedByRefer uuid.NullUUID `gorm:"type:char(36);"`
	CreatedBy      *User         `gorm:"->; foreignKey:CreatedByRefer; constraint:OnDelete:CASCADE;"`
	// Model structが含まれていたフィールド（例: CreatedAt, UpdatedAt, DeletedAtなど）
	// 必要に応じてここに明示的に定義してください。
	// CreatedAt time.Time
	// UpdatedAt time.Time
	// DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (*v16Group) TableName() string {
	return "groups" // テーブル名は変更されないことを明示
}

// v15Group はマイグレーション前のGroupモデルのテーブルカラムを反映します
// 提供されたv15Groupの定義を参考にします（リレーションシップは含まないことが多い）
type v16GroupOld struct {
	ID             uuid.UUID `gorm:"type:char(36);primaryKey"`
	Name           string    `gorm:"type:varchar(32);not null"`
	Description    string    `gorm:"type:TEXT"`
	IsTraqGroup    bool      `gorm:"not null;default:false"`
	JoinFreely     sql.NullBool
	TraqID         uuid.NullUUID `gorm:"default:null;uniqueIndex"`
	CreatedByRefer uuid.NullUUID `gorm:"type:char(36);"`
	// Model structが含まれていたフィールド（例: CreatedAt, UpdatedAt, DeletedAtなど）
	// 必要に応じてここに明示的に定義してください。
}

func (*v16GroupOld) TableName() string {
	return "groups" // テーブル名は変更されないことを明示
}

// v15GroupMember はマイグレーション前のGroupMember結合テーブル構造体を反映します
// 提供されたv15GroupMemberの定義を参考にします
type v16GroupMember struct {
	UserID  uuid.UUID `gorm:"type:char(36);primaryKey"`
	GroupID uuid.UUID `gorm:"type:char(36);primaryKey"`
	// Model structが含まれていたフィールド（例: CreatedAt, UpdatedAtなど）
	// もし結合テーブルにこれらのカラムがあった場合、ロールバックの際に必要になるため定義してください
	// CreatedAt time.Time
	// UpdatedAt time.Time
}

func (*v16GroupMember) TableName() string {
	return "group_members" // テーブル名は変更されないことを明示
}

// v15GroupAdmin はマイグレーション前のGroupAdmin結合テーブル構造体を反映します
// 提供されたv15GroupAdminの定義を参考にします
type v16GroupAdmin struct {
	UserID  uuid.UUID `gorm:"type:char(36);primaryKey"`
	GroupID uuid.UUID `gorm:"type:char(36);primaryKey"`
	// Model structが含まれていたフィールド（例: CreatedAt, UpdatedAtなど）
	// もし結合テーブルにこれらのカラムがあった場合、ロールバックの際に必要になるため定義してください
	// CreatedAt time.Time
	// UpdatedAt time.Time
}

func (*v16GroupAdmin) TableName() string {
	return "group_admins" // テーブル名は変更されないことを明示
}

// v16 は many2many リレーションシップへの変更を処理するマイグレーションです
func v16() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "16", // 前のマイグレーションIDの次のユニークなIDを設定

		Migrate: func(tx *gorm.DB) error {
			// GORMに新しいモデル定義に基づいてデータベースをマイグレーションさせます。
			// many2manyタグは既存の'group_member'および'group_admin'テーブルを使用するようにGORMに指示します。
			// テーブルが存在しない場合は作成を試みますが、このケースでは既に存在すると想定されます。
			// Userモデルもリレーションシップの一部として含めます。
			return tx.AutoMigrate(&v16Group{}, &User{})
			// AutoMigrateは、既存のテーブルやカラムに対して破壊的な変更（カラムの削除など）を
			// デフォルトでは行いません。新しいカラムやテーブルを追加するのに安全です。
			// この場合、主にGORMの内部的なリレーションシップマッピングを更新します。
		},

		Rollback: func(tx *gorm.DB) error {
			// マイグレーション前のモデル定義に基づいてデータベースをロールバックさせます。
			// これにより、GORMの内部的なリレーションシップマッピングが、
			// 明示的なGroupMember/GroupAdmin構造体を使用する状態に戻ります。
			// ロールバックでは、v15のGroup、GroupMember、GroupAdmin構造体を使用してAutoMigrateを行います。
			return tx.AutoMigrate(&v16GroupOld{}, &v16GroupMember{}, &v16GroupAdmin{})
			// 注意: AutoMigrateは、ロールバックの場合でも基本的にはテーブルやカラムを追加/修正する挙動です。
			// このケースではスキーマが大きく変わらないためこれで十分ですが、
			// 複雑なスキーマ変更のロールバックではDROP TABLE/COLUMNなどの明示的なSQLが必要になる場合があります。
		},
	}
}
