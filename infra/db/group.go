package db

import (
	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"gorm.io/gorm"
)

func groupFullPreload(tx *gorm.DB) *gorm.DB {
	return tx.Preload("Members").Preload("Admins")
}

type writeGroupParams struct {
	domain.WriteGroupParams
	CreatedBy uuid.UUID
}

func (repo *GormRepository) GetGroup(groupID uuid.UUID) (*Group, error) {
	return getGroup(repo.db, groupID)
}

func createGroup(db *gorm.DB, groupParams writeGroupParams) (*Group, error) {
	group := ConvertwriteGroupParamsToGroup(groupParams)
	err := db.Create(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func getGroup(db *gorm.DB, groupID uuid.UUID) (*Group, error) {
	group := Group{
		ID: groupID,
	}
	cmd := groupFullPreload(db)
	err := cmd.Take(&group).Error
	return &group, err
}

func getUserBelongingGroupIDs(db *gorm.DB, userID uuid.UUID) ([]uuid.UUID, error) {
	groupMembers := make([]*GroupMember, 0)

	err := db.Where("user_id = ?", userID).Find(&groupMembers).Error
	return ConvertSlicePointerGroupMemberToSliceuuidUUID(groupMembers), err
}
