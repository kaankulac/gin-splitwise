package model

import (
	"encoding/json"
	"fmt"
	"time"
)

type Group struct {
	ID uint `gorm:"column:id"`
	Title string `gorm:"column:title"`
	Description string `gorm:"column:description"`
	Image string `gorm:"column:image"`
	OwnerId uint `gorm:"column:owner_id"`
	TotalExpense uint `gorm:"column:total_expense"`
	TotalMember uint `gorm:"column:total_member"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

type GroupMember struct {
	GroupId uint `gorm:"column:group_id;primaryKey"`
	UserId uint `gorm:"column:user_id;primaryKey"`
	TotalDebt uint `gorm:"column:total_debt"`
	TotalCredit uint `gorm:"column:total_credit"`
}

func (g Group) String() string {
	return fmt.Sprintf("Group{ID:%d,title:%s,description:%s,image:%s,ownerId:%d,totalExpense:%d,totalMember:%d, CreatedAt: %v,UpdatedAt: %v}", g.ID, g.Title, g.Description, g.Image, g.OwnerId, g.TotalExpense, g.TotalMember, g.CreatedAt, g.UpdatedAt)
}

func (g *Group) UnmarshakJSON(b []byte) error {
	var tmp struct {
		ID uint `json:"id"`
		Title string `json:"title"`
		Description string `json:"description"`
		Image string `json:"image"`
		OwnerId uint `json:"owner_id"`
		TotalExpense uint `json:"total_expense"`
		TotalMember uint `json:"total_member"`
	}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}
	g.ID = tmp.ID
	g.Title = tmp.Title
	g.Description = tmp.Description
	g.Image = tmp.Image
	g.OwnerId = tmp.OwnerId
	g.TotalExpense = tmp.TotalExpense
	g.TotalMember = tmp.TotalMember
	return nil
}