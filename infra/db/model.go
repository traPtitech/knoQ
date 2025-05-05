package db

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

var tables = []interface{}{
	User{},
	Group{},
	Tag{},
	Room{},
	Event{},
	EventTag{}, // Eventより下にないと、overrideされる
	EventAttendee{},
}

// TODO: 本当に ProviderName を持っておく必要があるのか?
// 全て TraQIssuerName になるようなコードしか書いていないので
// YAGNI 的には削除すべきだと思うが
// 特に (repo *Repository) mergeDBUserandTraQUser(dbUser *db.User, traqUser *traq.User) (*domain.User, error)
// では traQIssuerNmae 以外に対しては error を返すような実装になっておりそれはもうこのサービスが
// traQ のみを対象にしていることと同値ではという
type User struct {
	ID uuid.UUID `gorm:"type:char(36); primaryKey"`
	// TODO: privileged に名前を変える
	Privileged   bool `gorm:"not null"`
	State        int
	IcalSecret   string `gorm:"not null"`
	ProviderName string `gorm:"not null"`
	AccessToken  string `gorm:"type:varbinary(64)"`
}

type Room struct {
	gorm.Model
	ID        uuid.UUID `gorm:"type:char(36);primaryKey"`
	Place     string    `gorm:"type:varchar(32);"`
	Verified  bool
	TimeStart time.Time `gorm:"type:DATETIME; index"`
	TimeEnd   time.Time `gorm:"type:DATETIME; index"`

	CreatedByRefer uuid.UUID `gorm:"type:char(36);" cvt:"CreatedBy, <-"`
	CreatedBy      User      `gorm:"->; foreignKey:CreatedByRefer; constraint:OnDelete:CASCADE;" cvt:"->"`

	Admins []*User `gorm:"many2many:room_admin;"`
	Events []Event `gorm:"->; constraint:-"` // readOnly
}

// Group is user group
type Group struct {
	gorm.Model
	ID          uuid.UUID     `gorm:"type:char(36);primaryKey"`
	Name        string        `gorm:"type:varchar(32);not null"`
	Description string        `gorm:"type:TEXT"`
	IsTraqGroup bool          `gorm:"not null"`
	JoinFreely  sql.NullBool  `gorm:""`
	TraqID      uuid.NullUUID `gorm:""`

	CreatedByRefer uuid.NullUUID `gorm:"type:char(36);" cvt:"CreatedBy, <-"`
	CreatedBy      *User         `gorm:"->; foreignKey:CreatedByRefer; constraint:OnDelete:CASCADE;" cvt:"->"`

	Members []*User `gorm:"many2many:group_member;"`
	Admins  []*User `gorm:"many2many:group_admin;"`
}

type Tag struct {
	gorm.Model
	ID   uuid.UUID `gorm:"type:char(36);primaryKey"`
	Name string    `gorm:"unique; type:varchar(16) binary"`
}

// EventTag is
type EventTag struct {
	gorm.Model
	TagID   uuid.UUID `gorm:"type:char(36); primaryKey" cvt:"ID"`
	EventID uuid.UUID `gorm:"type:char(36); primaryKey"`
	Event   *Event    `gorm:"->; foreignKey:EventID; constraint:OnDelete:CASCADE;"`
	Tag     *Tag      `gorm:"foreignKey:TagID; constraint:OnDelete:CASCADE;" cvt:"write:Name"`
	Locked  bool
}

type EventAttendee struct {
	UserID   uuid.UUID `gorm:"type:char(36); primaryKey"`
	EventID  uuid.UUID `gorm:"type:char(36); primaryKey"`
	User     User      `gorm:"->; foreignKey:UserID; constraint:OnDelete:CASCADE;" cvt:"->"`
	Schedule int
}

// Event is event for gorm
type Event struct {
	gorm.Model
	ID            uuid.UUID `gorm:"type:char(36); primaryKey"`
	Name          string    `gorm:"type:varchar(32); not null"`
	Description   string    `gorm:"type:TEXT"`
	TimeStart     time.Time `gorm:"type:DATETIME; index"`
	TimeEnd       time.Time `gorm:"type:DATETIME; index"`
	AllowTogether bool
	Open          bool

	CreatedByRefer uuid.UUID `gorm:"type:char(36); not null" cvt:"CreatedBy, <-"`
	CreatedBy      User      `gorm:"->; foreignKey:CreatedByRefer; constraint:OnDelete:CASCADE;" cvt:"->"`

	GroupID uuid.UUID `gorm:"type:char(36); not null; index"`
	Group   Group     `gorm:"->; foreignKey:GroupID; constraint:-"`

	RoomID uuid.UUID `gorm:"type:char(36); not null; index"`
	Room   Room      `gorm:"foreignKey:RoomID; constraint:OnDelete:CASCADE;" cvt:"write:Place"`

	Admins    []*User `gorm:"many2many:event_admin"`
	Attendees []EventAttendee
	Tags      []*EventTag `gorm:"foreignKey:EventID;references:ID;constraint:OnDelete:CASCADE;"`
}
