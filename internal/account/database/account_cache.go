package database

import (
	"context"
	"fmt"
	"gin-splitwise/internal/account/model"
	"gin-splitwise/internal/cache"
	"gin-splitwise/internal/metric"
)

var _ AccountDB = (*accountCacheDB)(nil)

const (
	cacheKeyUserByEmail = "user_by_email"
)

type accountCacheDB struct {
	cacher cache.Cacher
	mp *metric.MetricsProvider
	delegate AccountDB
}

func newAccountCacheDB(cacher cache.Cacher, mp *metric.MetricsProvider, delegate AccountDB) AccountDB {
	return &accountCacheDB{
		cacher: cacher,
		mp: mp,
		delegate: delegate,
	}
}

func (ac *accountCacheDB) Save(ctx context.Context, account *model.Account) error {
	if err := ac.delegate.Save(ctx, account); err != nil {
		return err
	}
	key := ac.userByEmailCacheKey(account.Email)
	ac.cacher.Set(ctx, key, account)
	return nil
}

func (ac *accountCacheDB) Update(ctx context.Context, email string, account *model.Account) error {
	if err := ac.delegate.Update(ctx, email, account); err != nil {
		return err
	}
	key := ac.userByEmailCacheKey(email)
	if exists, _ := ac.cacher.Exists(ctx, key); exists {
		find, err := ac.delegate.FindByEmail(ctx, email)
		if err == nil {
			ac.cacher.Set(ctx, key, find)
		}
	}
	return nil
}

func (ac *accountCacheDB) FindByEmail(ctx context.Context, email string) (*model.Account, error) {
	if cache.IsCacheSkip(ctx) {
		return ac.delegate.FindByEmail(ctx, email)
	}

	var (
		item model.Account
		key = ac.userByEmailCacheKey(email)
		cacheHit = true
	)
	err := ac.cacher.Fetch(ctx, key, &item, func() (interface{}, error) {
		cacheHit = false
		account, err := ac.delegate.FindByEmail(ctx, email)
		if err != nil {
			return nil ,err
		}
		account.Password = ""
		return account, nil
	})
	if err != nil {
		return nil, err
	}
	ac.mp.RecordCache(cacheKeyUserByEmail, cacheHit)
	return &item, nil
}

func (ac *accountCacheDB) userByEmailCacheKey(email string) string {
	return fmt.Sprintf("%s.%s", cacheKeyUserByEmail, email)
}