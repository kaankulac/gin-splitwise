package database

import (
	"context"
	"gin-splitwise/internal/account/model"
	"gin-splitwise/internal/cache"
	"gin-splitwise/internal/database"
	"gin-splitwise/internal/metric"
	"gin-splitwise/pkg/logging"

	"gorm.io/gorm"
)

type AccountDB interface {
	Save(ctx context.Context, account *model.Account) error

	Update(ctx context.Context, email string, account *model.Account) error

	FindByEmail(ctx context.Context, email string) (*model.Account, error)
}

func NewAccountDB(db *gorm.DB, cacher cache.Cacher, mp *metric.MetricsProvider) AccountDB {
	if cacher == nil {
		return &accountDB{db: db}
	}
	return newAccountCacheDB(cacher, mp, &accountDB{db: db})
}

type accountDB struct {
	db *gorm.DB
}

func (a *accountDB) Save(ctx context.Context, account *model.Account) error {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, a.db)
	logger.Debugw("account.db.Save", "account", account)

	if err := db.WithContext(ctx).Create(account).Error; err != nil {
		logger.Error("account.db.Save failed to save", "err", err)
		if database.IsKeyConflictError(err) {
			return database.ErrKeyConflict
		}
		return err
	}
	return nil
}

func (a *accountDB) Update(ctx context.Context, email string, account *model.Account) error {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, a.db)
	logger.Debugw("account.db.Update", "account", account)

	fields := make(map[string]interface{})
	if account.Username != "" {
		fields["username"] = account.Username
	}
	if account.Password != "" {
		fields["password"] = account.Password
	}
	if account.Image != "" {
		fields["image"] = account.Image
	}

	chain := db.WithContext(ctx).
		Model(&model.Account{}).
		Where("email = ?", email).
		UpdateColumns(fields)
	if chain.Error != nil {
		logger.Error("account.db.Update failed to update", "err", chain.Error)
		return chain.Error
	}
	if chain.RowsAffected == 0 {
		return database.ErrNotFound
	}
	return nil
}

func (a *accountDB) FindByEmail(ctx context.Context, email string) (*model.Account, error) {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, a.db)
	logger.Debugw("account.db.FindByEmail", "email", email)

	var acc model.Account
	if err := db.WithContext(ctx).Where("email = ?", email).First(&acc).Error; err != nil {
		logger.Error("account.db.FindByEmail failed to find", "err", err)
		if database.IsRecordNotFoundErr(err) {
			return nil, database.ErrNotFound
		}
		return nil, err
	}
	return &acc, nil
}