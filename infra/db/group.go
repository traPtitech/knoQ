package db

import (
	"github.com/gofrs/uuid"
	"github.com/traPtitech/knoQ/domain"
	"gorm.io/gorm"
)

func groupFullPreload(tx *gorm.DB) *gorm.DB {
	return tx.Preload("Members").Preload("Admins").Preload("CreatedBy")
}

type WriteGroupParams struct {
	domain.WriteGroupParams
	CreatedBy uuid.UUID
}

func (repo *GormRepository) CreateGroup(params WriteGroupParams) (*Group, error) {
	g, err := createGroup(repo.db, params)
	return g, defaultErrorHandling(err)
}

func (repo *GormRepository) UpdateGroup(groupID uuid.UUID, params WriteGroupParams) (*Group, error) {
	g, err := updateGroup(repo.db, groupID, params)
	return g, defaultErrorHandling(err)
}

func (repo *GormRepository) AddMemberToGroup(groupID, userID uuid.UUID) error {
	err := addMemberToGroup(repo.db, groupID, userID)
	return defaultErrorHandling(err)
}

func (repo *GormRepository) DeleteGroup(groupID uuid.UUID) error {
	err := deleteGroup(repo.db, groupID)
	return defaultErrorHandling(err)
}

func (repo *GormRepository) DeleteMemberOfGroup(groupID, userID uuid.UUID) error {
	err := deleteMemberOfGroup(repo.db, groupID, userID)
	return defaultErrorHandling(err)
}

func (repo *GormRepository) GetGroup(groupID uuid.UUID) (*Group, error) {
	gs, err := getGroup(groupFullPreload(repo.db), groupID)
	return gs, defaultErrorHandling(err)
}

func (repo *GormRepository) GetAllGroups() ([]*Group, error) {
	gs, err := getAllGroups(groupFullPreload(repo.db))
	return gs, defaultErrorHandling(err)
}

func (repo *GormRepository) GetUserBelongingGroupIDs(userID uuid.UUID) ([]uuid.UUID, error) {
	ids, err := getUserBelongingGroupIDs(repo.db, userID)
	return ids, defaultErrorHandling(err)
}

func createGroup(db *gorm.DB, groupParams WriteGroupParams) (*Group, error) {
	group := ConvWriteGroupParamsToGroup(groupParams)
	err := db.Create(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func updateGroup(db *gorm.DB, groupID uuid.UUID, params WriteGroupParams) (*Group, error) {
	group := ConvWriteGroupParamsToGroup(params)
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

func getAllGroups(db *gorm.DB) ([]*Group, error) {
	groups := make([]*Group, 0)
	err := db.Find(&groups).Error
	return groups, err
}

func getUserBelongingGroupIDs(db *gorm.DB, userID uuid.UUID) ([]uuid.UUID, error) {
	groupMembers := make([]*GroupMember, 0)

	err := db.Where("user_id = ?", userID).Find(&groupMembers).Error
	return ConvSPGroupMemberToSuuidUUID(groupMembers), err
}
