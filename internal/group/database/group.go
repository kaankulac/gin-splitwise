package database

import (
	"context"
	"gin-splitwise/internal/database"
	"gin-splitwise/internal/group/model"
	"gin-splitwise/pkg/logging"

	"gorm.io/gorm"
)

type GroupDB interface {
	Save(ctx context.Context, group *model.Group) error

	Update(ctx context.Context, id uint, group *model.Group) error

	FindByID(ctx context.Context, id uint) (*model.Group, error)

	Delete(ctx context.Context, id uint) error

	CheckOwner(ctx context.Context, groupID uint, userID uint) (bool, error)

	IncrementMemberCount(ctx context.Context, id uint) error

	DecrementMemberCount(ctx context.Context, id uint) error

	IncrementTotalExpenseCount(ctx context.Context, id uint, amount uint) error

	DecrementTotalExpenseCount(ctx context.Context, id uint, amount uint) error
}

func NewGroupDB(db *gorm.DB) GroupDB {
	return &groupDB{db: db}
}

type groupDB struct {
	db *gorm.DB
}

func (g *groupDB) Save(ctx context.Context, group *model.Group) error {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, g.db)
	logger.Debugw("group.db.Save", "group", group)

	if err := db.WithContext(ctx).Create(group).Error; err != nil {
		logger.Error("group.db.Save failed to save", "err", err)
		return err
	}
	return nil
}

func (g *groupDB) Update(ctx context.Context, id uint, group *model.Group) error {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, g.db)
	logger.Debugw("group.db.Update", "group", group)

	fields := make(map[string]interface{})
	if group.Title != "" {
		fields["title"] = group.Title
	}
	if group.Description != "" {
		fields["description"] = group.Description
	}
	if group.Image != "" {
		fields["image"] = group.Image
	}

	chain := db.WithContext(ctx).
		Model(&model.Group{}).
		Where("id = ?", id).
		UpdateColumns(fields)
	if chain.Error != nil {
		logger.Error("group.db.Update failed to update", "err", chain.Error)
		return chain.Error
	}
	if chain.RowsAffected == 0 {
		return database.ErrNotFound
	}
	return nil
}

func (g *groupDB) FindByID(ctx context.Context, id uint) (*model.Group, error) {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, g.db)
	logger.Debugw("group.db.FindByID", "id", id)

	var group model.Group
	if err := db.WithContext(ctx).Where("id = ?", id).First(&group).Error; err != nil {
		if database.IsRecordNotFoundErr(err) {
			return nil, database.ErrNotFound
		}
		return nil, err
	}
	return &group, nil
}

func (g *groupDB) Delete(ctx context.Context, id uint) error {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, g.db)
	logger.Debugw("group.db.Delete", "id", id)

	chain := db.WithContext(ctx).Where("id = ?", id).Delete(&model.Group{})
	if chain.Error != nil {
		logger.Error("group.db.Delete failed to delete", "err", chain.Error)
		return chain.Error
	}
	if chain.RowsAffected == 0 {
		return database.ErrNotFound
	}
	return nil
}

func (g *groupDB) CheckOwner(ctx context.Context, groupID uint, userID uint) (bool, error) {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, g.db)
	logger.Debugw("group.db.CheckOwner", "groupID", groupID, "userID", userID)

	var count int64
	if err := db.WithContext(ctx).Model(&model.Group{}).Where("id = ? AND owner_id = ?", groupID, userID).Count(&count).Error; err != nil {
		logger.Error("group.db.CheckOwner failed to check", "err", err)
		return false, err
	}
	return count > 0, nil
}

func (g *groupDB) IncrementMemberCount(ctx context.Context, id uint) error {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, g.db)
	logger.Debugw("group.db.IncrementMemberCount", "id", id)

	chain := db.WithContext(ctx).
		Model(&model.Group{}).
		Where("id = ?", id).
		Update("total_member", gorm.Expr("total_member + ?", 1))
	if chain.Error != nil {
		logger.Error("group.db.IncrementMemberCount failed to update", "err", chain.Error)
		return chain.Error
	}
	if chain.RowsAffected == 0 {
		return database.ErrNotFound
	}
	return nil
}

func (g *groupDB) DecrementMemberCount(ctx context.Context, id uint) error {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, g.db)
	logger.Debugw("group.db.DecrementMemberCount", "id", id)

	chain := db.WithContext(ctx).
		Model(&model.Group{}).
		Where("id = ?", id).
		Update("total_member", gorm.Expr("total_member - ?", 1))
	if chain.Error != nil {
		logger.Error("group.db.DecrementMemberCount failed to update", "err", chain.Error)
		return chain.Error
	}
	if chain.RowsAffected == 0 {
		return database.ErrNotFound
	}
	return nil
}

func (g *groupDB) IncrementTotalExpenseCount(ctx context.Context, id uint, amount uint) error {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, g.db)
	logger.Debugw("group.db.IncrementTotalExpenseCount", "id", id)

	chain := db.WithContext(ctx).
		Model(&model.Group{}).
		Where("id = ?", id).
		Update("total_expense", gorm.Expr("total_expense + ?", amount))
	if chain.Error != nil {
		logger.Error("group.db.IncrementTotalExpenseCount failed to update", "err", chain.Error)
		return chain.Error
	}
	if chain.RowsAffected == 0 {
		return database.ErrNotFound
	}
	return nil
}

func (g *groupDB) DecrementTotalExpenseCount(ctx context.Context, id uint, amount uint) error {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, g.db)
	logger.Debugw("group.db.DecrementTotalExpenseCount", "id", id)

	chain := db.WithContext(ctx).
		Model(&model.Group{}).
		Where("id = ?", id).
		Update("total_expense", gorm.Expr("total_expense - ?", amount))
	if chain.Error != nil {
		logger.Error("group.db.DecrementTotalExpenseCount failed to update", "err", chain.Error)
		return chain.Error
	}
	if chain.RowsAffected == 0 {
		return database.ErrNotFound
	}
	return nil
}
