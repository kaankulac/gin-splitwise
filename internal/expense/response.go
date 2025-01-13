package expense

import (
	"gin-splitwise/internal/expense/model"
)

type ExpenseResponse struct {
	Expense Expense `json:"expense"`
}

type Expense struct {
	ID            uint        `json:"id"`
	GroupId       uint        `json:"group_id"`
	User          ExpenseUser `json:"user"`
	TotalAmount   int         `json:"total_amount"`
	PerUserAmount int         `json:"per_user_amount"`
	CreatedAt     string      `json:"created_at"`
}

type UserBalanceResponse struct {
	UserBalance UserBalance `json:"user_balance"`
}

type UserBalance struct {
	User         ExpenseUser `json:"user_id"`
	CounterParty ExpenseUser `json:"counter_party_id"`
	Balance      int         `json:"balance"`
}

type ExpenseUser struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

func NewExpenseResponse(expense *model.Expense, user ExpenseUser) *ExpenseResponse {
	return &ExpenseResponse{
		Expense: Expense{
			ID:            expense.ID,
			GroupId:       expense.GroupId,
			User:          user,
			TotalAmount:   int(expense.TotalAmount),
			PerUserAmount: int(expense.PerUserAmount),
			CreatedAt:     expense.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	}
}

func NewExpensesResponse(expenses []model.Expense) []Expense {
	var response []Expense
	for _, expense := range expenses {
		response = append(response, Expense{
			ID:            expense.ID,
			GroupId:       expense.GroupId,
			User:          ExpenseUser{ID: expense.UserId},
			TotalAmount:   int(expense.TotalAmount),
			PerUserAmount: int(expense.PerUserAmount),
			CreatedAt:     expense.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return response
}

func NewUserBalanceResponse(userBalance *model.UserBalance, user ExpenseUser, counterParty ExpenseUser) *UserBalanceResponse {
	return &UserBalanceResponse{
		UserBalance: UserBalance{
			User:         user,
			CounterParty: counterParty,
			Balance:      int(userBalance.Balance),
		},
	}
}
