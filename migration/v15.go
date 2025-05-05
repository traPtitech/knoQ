package migration

import (
	"context"
	"database/sql"
	"errors"
	"os"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/gofrs/uuid"
	"github.com/samber/lo"
	"github.com/traPtitech/go-traq"
	"gorm.io/gorm"
)

type v15Group struct {
	ID             uuid.UUID `gorm:"type:char(36);primaryKey"`
	Name           string
	Description    string
	IsTraqGroup    bool `gorm:"not null;default:false"`
	JoinFreely     sql.NullBool
	TraqID         uuid.NullUUID `gorm:"default:null;uniqueIndex"`
	CreatedByRefer uuid.NullUUID `gorm:"type:char(36);" cvt:"CreatedBy, <-"`
}

func (*v15Group) TableName() string {
	return "groups"
}

type v15GroupMember struct {
	UserID  uuid.UUID `gorm:"type:char(36);primaryKey"`
	GroupID uuid.UUID `gorm:"type:char(36);primaryKey"`
}

func (*v15GroupMember) TableName() string {
	return "group_members"
}

type v15GroupAdmin struct {
	UserID  uuid.UUID `gorm:"type:char(36);primaryKey"`
	GroupID uuid.UUID `gorm:"type:char(36);primaryKey"`
}

func (*v15GroupAdmin) TableName() string {
	return "group_admins"
}

type v15Event struct {
	ID      uuid.UUID `gorm:"type:char(36);primaryKey"`
	GroupID uuid.UUID `gorm:"type:char(36);not null"`
}

func (*v15Event) TableName() string {
	return "events"
}

func newTraqAPIClient() ([]traq.UserGroup, error) {
	accessToken := os.Getenv("TRAQ_ACCESSTOKEN")
	if accessToken == "" {
		return nil, errors.New("")
	}
	ctx := context.WithValue(context.TODO(), traq.ContextAccessToken, accessToken)
	apiClient := traq.NewAPIClient(traq.NewConfiguration())
	gs, resp, err := apiClient.GroupApi.GetUserGroups(ctx).Execute()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, errors.New("")
	}

	return gs, nil
}

func v15() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "15",
		Migrate: func(tx *gorm.DB) error {
			groupIDs := make([]uuid.UUID, 0)
			if err := tx.Model(&v15Group{}).Pluck("id", &groupIDs).Error; err != nil {
				return err
			}
			groupSet := make(map[uuid.UUID]struct{}, 0)
			for _, id := range groupIDs {
				groupSet[id] = struct{}{}
			}

			// Add columns
			if err := tx.Migrator().AddColumn(&v15Group{}, "IsTraqGroup"); err != nil {
				return err
			}
			if err := tx.Migrator().AlterColumn(&v15Group{}, "JoinFreely"); err != nil {
				return err
			}
			if err := tx.Migrator().AlterColumn(&v15Group{}, "CreatedByRefer"); err != nil {
				return err
			}
			if err := tx.Migrator().AddColumn(&v15Group{}, "TraqID"); err != nil {
				return err
			}

			// Map[oldExternalID] -> newGeneratedUUID
			idMap := map[uuid.UUID]uuid.UUID{}

			tgs, err := newTraqAPIClient()
			if err != nil {
				return err
			}

			traqGroups := lo.Map(tgs, func(tg traq.UserGroup, _ int) *v15Group {
				tid := uuid.Must(uuid.FromString(tg.Id))
				newID := tid
				if _, ok := groupSet[tid]; ok {
					newID = uuid.Must(uuid.NewV4())
					idMap[tid] = newID
				}

				traqID := uuid.NullUUID{}
				err := traqID.Scan(tid)
				if err != nil {
					panic(err.Error())
				}

				return &v15Group{
					ID:          newID,
					Name:        tg.Name,
					Description: tg.Description,
					JoinFreely:  sql.NullBool{},
					IsTraqGroup: true,
					TraqID:      traqID,
				}
			})

			// 1. Insert external groups with new UUIDs
			if err := tx.Create(&traqGroups).Error; err != nil {
				return err
			}

			// 2. Rewrite GroupID in dependent tables
			for oldID, newID := range idMap {
				// group_members
				if err := tx.Model(&v15GroupMember{}).Where("group_id = ?", oldID).Update("group_id", newID).Error; err != nil {
					return err
				}
				// group_admins
				if err := tx.Model(&v15GroupAdmin{}).Where("group_id = ?", oldID).Update("group_id", newID).Error; err != nil {
					return err
				}
				// events
				if err := tx.Model(&v15Event{}).Where("group_id = ?", oldID).Update("group_id", newID).Error; err != nil {
					return err
				}
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			// 外部グループだけ削除
			return tx.Where("provider_name = ?", "traQ").Delete(&v15Group{}).Error
		},
	}
}
