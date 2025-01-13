package database

import (
	"context"

	"gorm.io/gorm"
)

type contextKey string

const dbKey = contextKey("db")

func WithDB(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, dbKey, db)
}

func FromContext(ctx context.Context, db *gorm.DB) *gorm.DB {
	if ctx == nil {
		return db
	}
	if stored, ok := ctx.Value(dbKey).(*gorm.DB); ok {
		return stored
	}
	return db
}