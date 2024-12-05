package shared

import (
	"gin-splitwise/internal/account/model"

	"github.com/gin-gonic/gin"
)

var identityKey = "id"

func CurrentUser(c *gin.Context) (*model.Account, bool) {
	data, ok := c.Get(identityKey)
	if !ok {
		return nil, false
	}
	acc, ok := data.(*model.Account)
	return acc, ok
}

func MustCurrentUser(c *gin.Context) *model.Account {
	acc, ok := CurrentUser(c)
	if ok {
		return acc
	}
	panic("no account in gin.Context")
}