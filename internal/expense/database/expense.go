package database

import (
	"context"
	"gin-splitwise/internal/database"
	"gin-splitwise/internal/expense/model"
	"gin-splitwise/pkg/logging"

	"gorm.io/gorm"
)

type ExpenseDB interface {
	Save(ctx context.Context, expense *model.Expense) error

	FindById(ctx context.Context, id uint) (*model.Expense, error) 

	Delete(ctx context.Context, id uint) error

	FindByGroup(ctx context.Context, groupId uint) ([]model.Expense, error)

	FindByUser(ctx context.Context, userId uint) ([]model.Expense, error)
}

type expenseDB struct {
	db *gorm.DB
}

func (e *expenseDB) Save(ctx context.Context, expense *model.Expense) error {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, e.db)
	logger.Debugw("expense.db.Save", "expense", expense)

	if err := db.WithContext(ctx).Create(expense).Error; err != nil {
		logger.Error("expense.db.Save failed to save", "err", err)
		return err
	}
	return nil
}

func (e *expenseDB) FindById(ctx context.Context, id uint) (*model.Expense, error) {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, e.db)
	logger.Debugw("expense.db.FindById", "id", id)

	var expense model.Expense
	if err := db.WithContext(ctx).Where("id = ?", id).First(&expense).Error; err != nil {
		if database.IsRecordNotFoundErr(err) {
			return nil, database.ErrNotFound
		}
		return nil, err
	}
	return &expense, nil
}

func (e *expenseDB) Delete(ctx context.Context, id uint) error {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, e.db)
	logger.Debugw("expense.db.Delete", "id", id)

	chain := db.WithContext(ctx).Where("id = ?", id).Delete(&model.Expense{})
	if chain.Error != nil {
		logger.Error("expense.db.Delete failed to delete", "err", chain.Error)
		return chain.Error
	}
	if chain.RowsAffected == 0 {
		return database.ErrNotFound
	}
	return nil
}

func (e *expenseDB) FindByGroup(ctx context.Context, groupId uint) ([]model.Expense, error) {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, e.db)
	logger.Debugw("expense.db.FindByGroup", "groupId", groupId)

	var expenses []model.Expense
	chain := db.WithContext(ctx).Where("group_id = ?", groupId).Find(&expenses)
	if chain.Error != nil {
		logger.Error("expense.db.FindByGroup failed to find", "err", chain.Error)
		return nil, chain.Error
	}
	return expenses, nil
}

func (e *expenseDB) FindByUser(ctx context.Context, userId uint) ([]model.Expense, error) {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, e.db)
	logger.Debugw("expense.db.FindByUser", "userId", userId)

	var expenses []model.Expense
	chain := db.WithContext(ctx).Where("user_id = ?", userId).Find(&expenses)
	if chain.Error != nil {
		logger.Error("expense.db.FindByUser failed to find", "err", chain.Error)
		return nil, chain.Error
	}
	return expenses, nil
}