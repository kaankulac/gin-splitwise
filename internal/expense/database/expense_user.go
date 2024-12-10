package database

import (
	"context"
	"gin-splitwise/internal/database"
	"gin-splitwise/internal/expense/model"
	"gin-splitwise/pkg/logging"

	"gorm.io/gorm"
)

type ExpenseUserDB interface {
	Save(ctx context.Context, expenseUser *model.ExpenseUser) error

	FindByUserId(ctx context.Context, userId uint) ([]model.ExpenseUser, error)

	FindByExpenseId(ctx context.Context, expenseId uint) ([]model.ExpenseUser, error)

	Delete(ctx context.Context, userId uint, expenseId uint) error

}

type expenseUserDB struct {
	db *gorm.DB
}

func (e *expenseUserDB) Save(ctx context.Context, expenseUser *model.ExpenseUser) error {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, e.db)
	logger.Debugw("expenseUser.db.Save", "expenseUser", expenseUser)

	if err := db.WithContext(ctx).Create(expenseUser).Error; err != nil {
		logger.Error("expenseUser.db.Save failed to save", "err", err)
		return err
	}
	return nil
}

func (e *expenseUserDB) FindByUserId(ctx context.Context, userId uint) ([]model.ExpenseUser, error) {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, e.db)
	logger.Debugw("expenseUser.db.FindByUserId", "userId", userId)

	var expenseUsers []model.ExpenseUser
	chain := db.WithContext(ctx).Where("user_id = ?", userId).Find(&expenseUsers)
	if chain.Error != nil {
		logger.Error("expenseUser.db.FindByUserId failed to find", "err", chain.Error)
		return nil, chain.Error
	}
	return expenseUsers, nil
}

func (e *expenseUserDB) FindByExpenseId(ctx context.Context, expenseId uint) ([]model.ExpenseUser, error) {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, e.db)
	logger.Debugw("expenseUser.db.FindByExpenseId", "expenseId", expenseId)

	var expenseUsers []model.ExpenseUser
	chain := db.WithContext(ctx).Where("expense_id = ?", expenseId).Find(&expenseUsers)
	if chain.Error != nil {
		logger.Error("expenseUser.db.FindByExpenseId failed to find", "err", chain.Error)
		return nil, chain.Error
	}
	return expenseUsers, nil
}

func (e *expenseUserDB) Delete(ctx context.Context, userId uint, expenseId uint) error {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, e.db)
	logger.Debugw("expenseUser.db.Delete", "userId", userId, "expenseId", expenseId)

	chain := db.WithContext(ctx).Where("user_id = ? AND expense_id = ?", userId, expenseId).Delete(&model.ExpenseUser{})
	if chain.Error != nil {
		logger.Error("expenseUser.db.Delete failed to delete", "err", chain.Error)
		return chain.Error
	}
	if chain.RowsAffected == 0 {
		return database.ErrNotFound
	}
	return nil
}