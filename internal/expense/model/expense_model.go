package model

import (
	"encoding/json"
	"fmt"
	"time"
)

type Expense struct {
	ID uint `gorm:"column:id"`
	GroupId uint `gorm:"column:group_id;primaryKey"`
	UserId uint `gorm:"column:user_id;primaryKey"`
	TotalAmount uint `gorm:"column:total_amount"`
	PerUserAmount uint `gorm:"column:per_user_amount"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

type ExpenseUser struct {
	UserId uint `gorm:"column:user_id;primaryKey"`
	ExpenseId uint `gorm:"column:expense_id;primaryKey"`
	Amount uint `gorm:"column:amount"`
}

func (e Expense) String() string {
	return fmt.Sprintf("Expense{GroupId:%d,UserId:%d,TotalAmount:%d,PerUserAmount:%d,CreatedAt:%v}", e.GroupId, e.UserId, e.TotalAmount, e.PerUserAmount, e.CreatedAt)
}

func (e *Expense) UnmarshalJSON(b []byte) error {
	var tmp struct {
		ID uint `json:"id"`
		GroupId uint `json:"group_id"`
		UserId uint `json:"user_id"`
		TotalAmount uint `json:"total_amount"`
		PerUserAmount uint `json:"per_user_amount"`
	}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}
	e.ID = tmp.ID
	e.GroupId = tmp.GroupId
	e.UserId = tmp.UserId
	e.TotalAmount = tmp.TotalAmount
	e.PerUserAmount = tmp.PerUserAmount
	return nil
}