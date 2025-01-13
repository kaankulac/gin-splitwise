package database

import (
	"context"
	"gin-splitwise/internal/database"
	"gin-splitwise/internal/expense/model"
	"gin-splitwise/pkg/logging"

	"gorm.io/gorm"
)

type UserBalanceDB interface {
	Save(ctx context.Context, userBalance *model.UserBalance) error

	FindByUsers(ctx context.Context, userId uint, counterPartyId uint) (*model.UserBalance, error)

	Delete(ctx context.Context, userId uint, counterPartyId uint) error

	IncrementBalance(ctx context.Context, userId uint, counterPartyId uint, amount uint) error

	DecrementBalance(ctx context.Context, userId uint, counterPartyId uint, amount uint) error
}

type userBalanceDB struct {
	db *gorm.DB
}

func (u *userBalanceDB) Save(ctx context.Context, userBalance *model.UserBalance) error {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, u.db)
	logger.Debugw("userBalance.db.Save", "userBalance", userBalance)

	if err := db.WithContext(ctx).Create(userBalance).Error; err != nil {
		logger.Error("userBalance.db.Save failed to save", "err", err)
		if database.IsKeyConflictError(err) {
			return database.ErrKeyConflict
		}
		return err
	}
	return nil
}

func (u *userBalanceDB) FindByUsers(ctx context.Context, userId uint, counterPartyId uint) (*model.UserBalance, error) {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, u.db)
	logger.Debugw("userBalance.db.FindByUsers", "userId", userId, "counterPartyId", counterPartyId)

	var userBalance model.UserBalance
	chain := db.WithContext(ctx).Where("user_id = ? AND counter_party_id = ?", userId, counterPartyId).Find(&userBalance)
	if chain.Error != nil {
		logger.Error("userBalance.db.FindByUsers failed to find", "err", chain.Error)
		return nil, chain.Error
	}
	return &userBalance, nil
}

func (u *userBalanceDB) Delete(ctx context.Context, userId uint, counterPartyId uint) error {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, u.db)
	logger.Debugw("userBalance.db.Delete", "userId", userId, "counterPartyId", counterPartyId)

	chain := db.WithContext(ctx).Where("user_id = ? AND counter_party_id = ?", userId, counterPartyId).Delete(&model.UserBalance{})
	if chain.Error != nil {
		logger.Error("userBalance.db.Delete failed to delete", "err", chain.Error)
		return chain.Error
	}
	if chain.RowsAffected == 0 {
		return database.ErrNotFound
	}
	return nil
}

func (u *userBalanceDB) IncrementBalance(ctx context.Context, userId uint, counterPartyId uint, amount uint) error {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, u.db)
	logger.Debugw("userBalance.db.IncrementBalance", "userId", userId, "counterPartyId", counterPartyId, "amount", amount)

	chain := db.WithContext(ctx).
		Model(&model.UserBalance{}).
		Where("user_id = ? AND counter_party_id = ?", userId, counterPartyId).
		Update("balance", gorm.Expr("balance + ?", amount))
	if chain.Error != nil {
		logger.Error("userBalance.db.IncrementBalance failed to update", "err", chain.Error)
		return chain.Error
	}
	if chain.RowsAffected == 0 {
		return database.ErrNotFound
	}
	chain = db.WithContext(ctx).
		Model(&model.UserBalance{}).
		Where("user_id = ? AND counter_party_id = ?", counterPartyId, userId).
		Update("balance", gorm.Expr("balance - ?", amount))
	if chain.Error != nil {
		logger.Error("userBalance.db.IncrementBalance failed to update", "err", chain.Error)
		return chain.Error
	}
	return nil
}

func (u *userBalanceDB) DecrementBalance(ctx context.Context, userId uint, counterPartyId uint, amount uint) error {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, u.db)
	logger.Debugw("userBalance.db.DecrementBalance", "userId", userId, "counterPartyId", counterPartyId, "amount", amount)

	chain := db.WithContext(ctx).
		Model(&model.UserBalance{}).
		Where("user_id = ? AND counter_party_id = ?", userId, counterPartyId).
		Update("balance", gorm.Expr("balance - ?", amount))
	if chain.Error != nil {
		logger.Error("userBalance.db.DecrementBalance failed to update", "err", chain.Error)
		return chain.Error
	}
	if chain.RowsAffected == 0 {
		return database.ErrNotFound
	}
	chain = db.WithContext(ctx).
		Where("user_id = ? AND counter_party_id = ?", counterPartyId, userId).
		Update("balance", gorm.Expr("balance + ?", amount))
	if chain.Error != nil {
		logger.Error("userBalance.db.DecrementBalance failed to delete", "err", chain.Error)
		return chain.Error
	}
	return nil
}
