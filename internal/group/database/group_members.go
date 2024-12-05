package database

import (
	"context"
	"gin-splitwise/internal/database"
	"gin-splitwise/internal/group/model"
	"gin-splitwise/pkg/logging"

	"gorm.io/gorm"
)

type GroupMemberDB interface {
	Add(ctx context.Context, member *model.GroupMember) error

	Remove(ctx context.Context, userId uint, groupId uint) error

	FindByGroupID(ctx context.Context, id uint) ([]model.GroupMember, error)

	FindByUserID(ctx context.Context, id uint) ([]model.GroupMember, error)
}

type groupMemberDB struct {
	db *gorm.DB
}

func (g *groupMemberDB) Add(ctx context.Context, member *model.GroupMember) error {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, g.db)
	logger.Debugw("group_member.db.Add", "member", member)

	if err := db.WithContext(ctx).Create(member).Error; err != nil {
		logger.Error("group_member.db.Add failed to save", "err", err)
		if database.IsKeyConflictError(err) {
			return database.ErrKeyConflict
		}
		return err
	}
	return nil
}

func (g *groupMemberDB) Remove(ctx context.Context, userId uint, groupId uint) error {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, g.db)
	logger.Debugw("group_member.db.Remove", "userId", userId)

	chain := db.WithContext(ctx).
		Where("userId = ? AND group_id = ?", userId, groupId).
		Delete(&model.GroupMember{})
	if chain.Error != nil {
		logger.Error("group_member.db.Remove failed to delete", "err", chain.Error)
		return chain.Error
	}
	if chain.RowsAffected == 0 {
		return database.ErrNotFound
	}
	return nil
}

func (g *groupMemberDB) FindByGroupID(ctx context.Context, id uint) ([]model.GroupMember, error) {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, g.db)
	logger.Debugw("group_member.db.FindByGroupID", "id", id)

	var members []model.GroupMember
	chain := db.WithContext(ctx).
		Where("group_id = ?", id).
		Find(members)
	if chain.Error != nil {
		logger.Error("group_member.db.FindByGroupID failed to find", "err", chain.Error)
		return nil, chain.Error
	}
	return members, nil
}

func (g *groupMemberDB) FindByUserID(ctx context.Context, id uint) ([]model.GroupMember, error) {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, g.db)
	logger.Debugw("group_member.db.FindByUserID", "id", id)

	var members []model.GroupMember
	chain := db.WithContext(ctx).
		Where("user_id = ?", id).
		Find(members)
	if chain.Error != nil {
		logger.Error("group_member.db.FindByUserID failed to find", "err", chain.Error)
		return nil, chain.Error
	}
	return members, nil
}