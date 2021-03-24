package db

import (
	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"gorm.io/gorm"
)

type writeGroupParams struct {
	domain.WriteGroupParams
	CreatedBy uuid.UUID
}

func createGroup(db *gorm.DB, groupParams writeGroupParams) (*Group, error) {
	group := ConvertwriteGroupParamsToGroup(groupParams)
	err := db.Create(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func getUserBelongingGroupIDs(db *gorm.DB, userID uuid.UUID) ([]uuid.UUID, error) {
	groupMembers := make([]*GroupMember, 0)

	err := db.Where("user_id = ?", userID).Find(&groupMembers).Error
	return ConvertSlicePointerGroupMemberToSliceuuidUUID(groupMembers), err
}
