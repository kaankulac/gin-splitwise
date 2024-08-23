package database

import (
	"errors"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

var (
	ErrNotFound = errors.New("record not found")
	ErrKeyConflict = errors.New("key conflict")
)

func IsRecordNotFoundErr(err error) bool {
	return err == gorm.ErrRecordNotFound || err == ErrNotFound
}

func IsKeyConflictError(err error) bool {
	if err == ErrKeyConflict {
		return true
	}
	switch err := err.(type) {
	case *pq.Error:
		if err.Code == "23505" {
			return true
		}
	}
	return false
}