package db

import (
	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"gorm.io/gorm"
)

func groupFullPreload(tx *gorm.DB) *gorm.DB {
	return tx.Preload("Members").Preload("Admins")
}

type WriteGroupParams struct {
	domain.WriteGroupParams
	CreatedBy uuid.UUID
}

func (repo *GormRepository) CreateGroup(params WriteGroupParams) (*Group, error) {
	return createGroup(repo.db, params)
}

func (repo *GormRepository) UpdateGroup(groupID uuid.UUID, params WriteGroupParams) (*Group, error) {
	return updateGroup(repo.db, groupID, params)
}

func (repo *GormRepository) AddMemberToGroup(db *gorm.DB, groupID, userID uuid.UUID) error {
	return addMemberToGroup(repo.db, groupID, userID)
}

func (repo *GormRepository) DeleteGroup(db *gorm.DB, groupID uuid.UUID) error {
	return deleteGroup(repo.db, groupID)
}

func (repo *GormRepository) DeleteMemberOfGroup(db *gorm.DB, groupID, userID uuid.UUID) error {
	return deleteMemberOfGroup(repo.db, groupID, userID)
}

func (repo *GormRepository) GetGroup(groupID uuid.UUID) (*Group, error) {
	return getGroup(groupFullPreload(repo.db), groupID)
}

func (repo *GormRepository) GetAllGroups(db *gorm.DB, groupID uuid.UUID) ([]*Group, error) {
	return getAllGroups(groupFullPreload(repo.db), groupID)
}

func (repo *GormRepository) GetUserBelongingGroupIDs(db *gorm.DB, userID uuid.UUID) ([]uuid.UUID, error) {
	return getUserBelongingGroupIDs(repo.db, userID)
}

func createGroup(db *gorm.DB, groupParams WriteGroupParams) (*Group, error) {
	group := ConvertWriteGroupParamsToGroup(groupParams)
	err := db.Create(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func updateGroup(db *gorm.DB, groupID uuid.UUID, params WriteGroupParams) (*Group, error) {
	group := ConvertWriteGroupParamsToGroup(params)
	group.ID = groupID
	err := db.Session(&gorm.Session{FullSaveAssociations: true}).Save(&group).Error
	return &group, err
}

func addMemberToGroup(db *gorm.DB, groupID, userID uuid.UUID) error {
	groupMember := GroupMember{
		GroupID: groupID,
		UserID:  userID,
	}
	return db.Create(&groupMember).Error
}

func deleteGroup(db *gorm.DB, groupID uuid.UUID) error {
	group := Group{
		ID: groupID,
	}
	return db.Delete(&group).Error
}

func deleteMemberOfGroup(db *gorm.DB, groupID, userID uuid.UUID) error {
	groupMember := GroupMember{
		GroupID: groupID,
		UserID:  userID,
	}
	return db.Delete(&groupMember).Error
}

func getGroup(db *gorm.DB, groupID uuid.UUID) (*Group, error) {
	group := Group{}
	err := db.Take(&group, groupID).Error
	return &group, err
}

func getAllGroups(db *gorm.DB, groupID uuid.UUID) ([]*Group, error) {
	groups := make([]*Group, 0)
	err := db.Find(&groups).Error
	return groups, err
}

func getUserBelongingGroupIDs(db *gorm.DB, userID uuid.UUID) ([]uuid.UUID, error) {
	groupMembers := make([]*GroupMember, 0)

	err := db.Where("user_id = ?", userID).Find(&groupMembers).Error
	return ConvertSlicePointerGroupMemberToSliceuuidUUID(groupMembers), err
}
