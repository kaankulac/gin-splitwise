package expense

import (
	"gin-splitwise/internal/config"
	expenseDB "gin-splitwise/internal/expense/database"
	"gin-splitwise/internal/expense/model"
	groupDB "gin-splitwise/internal/group/database"
	"gin-splitwise/internal/middleware"
	"gin-splitwise/internal/middleware/handler"
	"gin-splitwise/internal/shared"
	"gin-splitwise/pkg/logging"
	"gin-splitwise/pkg/validate"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	expenseDB     expenseDB.ExpenseDB
	expenseUserDB expenseDB.ExpenseUserDB
	userBalanceDB expenseDB.UserBalanceDB
	groupDB       groupDB.GroupDB
	groupMemberDB groupDB.GroupMemberDB
}

func (h *Handler) createExpense(c *gin.Context) {
	handler.HandleRequest(c, func(c *gin.Context) *handler.Response {
		logger := logging.FromContext(c)
		type RequestBody struct {
			Expense struct {
				GroupId     uint `json:"group_id" binding:"required,numeric"`
				TotalAmount uint `json:"total_amount" binding:"required,numeric"`
			}
		}
		var body RequestBody
		if err := c.ShouldBindJSON(&body); err != nil {
			logger.Errorw("expense.handler.createExpense failed to bind", "err", err)
			var details []*validate.ValidationErrDetail
			if vErrs, ok := err.(validator.ValidationErrors); ok {
				details = validate.ValidationErrorDetails(&body.Expense, "json", vErrs)
			}
			return handler.NewErrorResponse(http.StatusBadRequest, handler.InvalidBodyValue, "invalid expense payload in body", details)
		}
		currentUser := shared.MustCurrentUser(c)
		group, err := h.groupDB.FindByID(c, body.Expense.GroupId)
		if err != nil {
			logger.Errorw("expense.handler.createExpense failed to find group", "err", err)
			return handler.NewErrorResponse(http.StatusInternalServerError, handler.NotFoundEntity, "failed to find group", nil)
		}
		expense := model.Expense{
			GroupId:       body.Expense.GroupId,
			TotalAmount:   body.Expense.TotalAmount,
			UserId:        currentUser.ID,
			PerUserAmount: body.Expense.TotalAmount / group.TotalMember,
			HasPayment:    false,
		}
		err = h.expenseDB.Save(c, &expense)
		if err != nil {
			logger.Errorw("expense.handler.createExpense failed to save expense", "err", err)
			return handler.NewErrorResponse(http.StatusInternalServerError, handler.InternalServerError, "failed to save expense", nil)
		}

		members, error := h.groupMemberDB.FindByGroupID(c, body.Expense.GroupId)
		if error != nil {
			logger.Errorw("expense.handler.createExpense failed to find group members", "err", err)
			return handler.NewErrorResponse(http.StatusInternalServerError, handler.InternalServerError, "failed to find group members", nil)
		}

		for _, member := range members {
			expenseUser := model.ExpenseUser{
				ExpenseId: expense.ID,
				UserId:    member.UserId,
				Paid:      false,
			}
			err = h.expenseUserDB.Save(c, &expenseUser)
			if err != nil {
				logger.Errorw("expense.handler.createExpense failed to save expense user", "err", err)
				return handler.NewErrorResponse(http.StatusInternalServerError, handler.InternalServerError, "failed to save expense user", nil)
			}
			if member.UserId == currentUser.ID {
				continue
			}
			checkUserBalance, _ := h.userBalanceDB.FindByUsers(c, currentUser.ID, member.UserId)
			if checkUserBalance == nil {
				userBalance := model.UserBalance{
					UserId:         currentUser.ID,
					CounterPartyId: member.UserId,
					GroupId:        expense.GroupId,
					Balance:        expense.PerUserAmount,
				}
				counterUserBalance := model.UserBalance{
					UserId:         member.UserId,
					CounterPartyId: currentUser.ID,
					GroupId:        expense.GroupId,
					Balance:        -expense.PerUserAmount,
				}
				err := h.userBalanceDB.Save(c, &userBalance)
				if err != nil {
					logger.Errorw("expense.handler.createExpense failed to save user balance", "err", err)
					return handler.NewErrorResponse(http.StatusInternalServerError, handler.InternalServerError, "failed to save user balance", nil)
				}

				err = h.userBalanceDB.Save(c, &counterUserBalance)
				if err != nil {
					logger.Errorw("expense.handler.createExpense failed to save user balance", "err", err)
					return handler.NewErrorResponse(http.StatusInternalServerError, handler.InternalServerError, "failed to save user balance", nil)
				}
			} else {
				err := h.userBalanceDB.IncrementBalance(c, currentUser.ID, member.UserId, expense.PerUserAmount)
				if err != nil {
					logger.Errorw("expense.handler.createExpense failed to increment user balance", "err", err)
					return handler.NewErrorResponse(http.StatusInternalServerError, handler.InternalServerError, "failed to increment user balance", nil)
				}
			}
		}
		err = h.groupDB.IncrementTotalExpenseCount(c, body.Expense.GroupId, body.Expense.TotalAmount)
		if err != nil {
			logger.Errorw("expense.handler.createExpense failed to increment total expense count", "err", err)
			return handler.NewErrorResponse(http.StatusInternalServerError, handler.InternalServerError, "failed to increment total expense count", nil)
		}
		return handler.NewSuccessResponse(http.StatusCreated, NewExpenseResponse(&expense, ExpenseUser{
			ID:       currentUser.ID,
			Username: currentUser.Username,
		}))
	})
}

func (h *Handler) getExpensesByGroup(c *gin.Context) {
	handler.HandleRequest(c, func(c *gin.Context) *handler.Response {
		logger := logging.FromContext(c)
		type RequestUri struct {
			GroupId uint `uri:"groupId" binding:"required,numeric"`
		}
		var uri RequestUri
		if err := c.ShouldBindUri(&uri); err != nil {
			logger.Errorw("expense.handler.getExpensesByGroup failed to bind uri", "err", err)
			var details []*validate.ValidationErrDetail
			if vErrs, ok := err.(validator.ValidationErrors); ok {
				details = validate.ValidationErrorDetails(&uri, "uri", vErrs)
			}
			return handler.NewErrorResponse(http.StatusBadRequest, handler.InvalidBodyValue, "invalid group id in uri", details)
		}
		expenses, err := h.expenseDB.FindByGroup(c, uri.GroupId)
		if err != nil {
			logger.Errorw("expense.handler.getExpensesByGroup failed to find expenses", "err", err)
			return handler.NewErrorResponse(http.StatusInternalServerError, handler.InternalServerError, "failed to find expenses", nil)
		}

		return handler.NewSuccessResponse(http.StatusOK, NewExpensesResponse(expenses))
	})
}

func (h *Handler) deleteExpenseByGroup(c *gin.Context) {
	handler.HandleRequest(c, func(c *gin.Context) *handler.Response {
		logger := logging.FromContext(c)
		type RequestUri struct {
			ExpenseId uint `uri:"expenseId" binding:"required,numeric"`
		}
		var uri RequestUri
		if err := c.ShouldBindUri(&uri); err != nil {
			logger.Errorw("expense.handler.deleteExpenseByGroup failed to bind uri", "err", err)
			var details []*validate.ValidationErrDetail
			if vErrs, ok := err.(validator.ValidationErrors); ok {
				details = validate.ValidationErrorDetails(&uri, "uri", vErrs)
			}
			return handler.NewErrorResponse(http.StatusBadRequest, handler.InvalidBodyValue, "invalid expense id in uri", details)
		}
		currentUser := shared.MustCurrentUser(c)
		expense, err := h.expenseDB.FindById(c, uri.ExpenseId)
		if err != nil {
			logger.Errorw("expense.handler.deleteExpenseByGroup failed to find expense", "err", err)
			return handler.NewErrorResponse(http.StatusInternalServerError, handler.InternalServerError, "failed to find expense", nil)
		}
		if expense == nil {
			return handler.NewErrorResponse(http.StatusNotFound, handler.NotFoundEntity, "expense not found", nil)
		}
		if expense.UserId != currentUser.ID {
			return handler.NewErrorResponse(http.StatusForbidden, handler.Forbidden, "current user is not owner of the expense", nil)
		}
		if expense.HasPayment {
			return handler.NewErrorResponse(http.StatusForbidden, handler.Forbidden, "expense has payment", nil)
		}

		members, error := h.groupMemberDB.FindByGroupID(c, expense.GroupId)
		if error != nil {
			logger.Errorw("expense.handler.createExpense failed to find group members", "err", err)
			return handler.NewErrorResponse(http.StatusInternalServerError, handler.InternalServerError, "failed to find group members", nil)
		}
		for _, member := range members {
			err = h.expenseUserDB.Delete(c, member.UserId, expense.ID)
			if err != nil {
				logger.Errorw("expense.handler.deleteExpense failed to save expense user", "err", err)
				return handler.NewErrorResponse(http.StatusInternalServerError, handler.InternalServerError, "failed to delete expense user", nil)
			}
			if member.UserId == currentUser.ID {
				continue
			}
			checkUserBalance, _ := h.userBalanceDB.FindByUsers(c, currentUser.ID, member.UserId)
			if checkUserBalance == nil {
				userBalance := model.UserBalance{
					UserId:         currentUser.ID,
					CounterPartyId: member.UserId,
					GroupId:        expense.GroupId,
					Balance:        0,
				}
				counterUserBalance := model.UserBalance{
					UserId:         member.UserId,
					CounterPartyId: currentUser.ID,
					GroupId:        expense.GroupId,
					Balance:        0,
				}
				err := h.userBalanceDB.Save(c, &userBalance)
				if err != nil {
					logger.Errorw("expense.handler.createExpense failed to save user balance", "err", err)
					return handler.NewErrorResponse(http.StatusInternalServerError, handler.InternalServerError, "failed to save user balance", nil)
				}
				err = h.userBalanceDB.Save(c, &counterUserBalance)
				if err != nil {
					logger.Errorw("expense.handler.createExpense failed to save user balance", "err", err)
					return handler.NewErrorResponse(http.StatusInternalServerError, handler.InternalServerError, "failed to save user balance", nil)
				}
			} else {
				err := h.userBalanceDB.DecrementBalance(c, currentUser.ID, member.UserId, expense.PerUserAmount)
				if err != nil {
					logger.Errorw("expense.handler.createExpense failed to increment user balance", "err", err)
					return handler.NewErrorResponse(http.StatusInternalServerError, handler.InternalServerError, "failed to increment user balance", nil)
				}
			}
		}

		err = h.expenseDB.Delete(c, uri.ExpenseId)
		if err != nil {
			logger.Errorw("expense.handler.deleteExpenseByGroup failed to delete expense", "err", err)
			return handler.NewErrorResponse(http.StatusInternalServerError, handler.InternalServerError, "failed to delete expense", nil)
		}

		err = h.groupDB.DecrementTotalExpenseCount(c, expense.GroupId, expense.TotalAmount)
		if err != nil {
			logger.Errorw("expense.handler.deleteExpenseByGroup failed to decrement total expense count", "err", err)
			return handler.NewErrorResponse(http.StatusInternalServerError, handler.InternalServerError, "failed to decrement total expense count", nil)
		}

		return handler.NewSuccessResponse(http.StatusOK, nil)
	})
}

func (h *Handler) payExpense(c *gin.Context) {
	handler.HandleRequest(c, func(c *gin.Context) *handler.Response {
		logger := logging.FromContext(c)
		type RequestUri struct {
			ExpenseId uint `uri:"expenseId" binding:"required,numeric"`
		}
		var uri RequestUri
		if err := c.ShouldBindUri(&uri); err != nil {
			logger.Errorw("expense.handler.payExpense failed to bind uri", "err", err)
			var details []*validate.ValidationErrDetail
			if vErrs, ok := err.(validator.ValidationErrors); ok {
				details = validate.ValidationErrorDetails(&uri, "uri", vErrs)
			}
			return handler.NewErrorResponse(http.StatusBadRequest, handler.InvalidBodyValue, "invalid expense id in uri", details)
		}
		currentUser := shared.MustCurrentUser(c)
		expense, err := h.expenseDB.FindById(c, uri.ExpenseId)
		if err != nil {
			logger.Errorw("expense.handler.payExpense failed to find expense", "err", err)
			return handler.NewErrorResponse(http.StatusInternalServerError, handler.InternalServerError, "failed to find expense", nil)
		}
		if expense == nil {
			return handler.NewErrorResponse(http.StatusNotFound, handler.NotFoundEntity, "expense not found", nil)
		}
		if expense.UserId == currentUser.ID {
			return handler.NewErrorResponse(http.StatusForbidden, handler.Forbidden, "current user is owner of the expense", nil)
		}
		err = h.expenseUserDB.PayExpense(c, currentUser.ID, uri.ExpenseId)
		if err != nil {
			logger.Errorw("expense.handler.payExpense failed to pay expense", "err", err)
			return handler.NewErrorResponse(http.StatusInternalServerError, handler.InternalServerError, "failed to pay expense", nil)
		}

		err = h.userBalanceDB.DecrementBalance(c, expense.UserId, currentUser.ID, expense.PerUserAmount)
		if err != nil {
			logger.Errorw("expense.handler.payExpense failed to decrement user balance", "err", err)
			return handler.NewErrorResponse(http.StatusInternalServerError, handler.InternalServerError, "failed to decrement user balance", nil)
		}

		return handler.NewSuccessResponse(http.StatusOK, nil)
	})
}

func RouteV1(cfg *config.Config, h *Handler, r *gin.Engine, auth *jwt.GinJWTMiddleware) {
	v1 := r.Group("v1/api")
	v1.Use(middleware.RequestIDMiddleware(), middleware.TimeoutMiddleware(cfg.ServerConfig.WriteTimeout))

	// Authenticated Routes
	v1.Use(auth.MiddlewareFunc())
	{
		v1.GET("/expenses/:groupId", h.getExpensesByGroup)
		v1.POST("/expenses", h.createExpense)
		v1.DELETE("/expenses/:expenseId", h.deleteExpenseByGroup)
		v1.POST("/expenses/:expenseId/pay", h.payExpense)
	}
}

func NewHandler(groupDB groupDB.GroupDB, groupMemberDB groupDB.GroupMemberDB, expenseDB expenseDB.ExpenseDB, expenseUserDB expenseDB.ExpenseUserDB, userBalanceDB expenseDB.UserBalanceDB) *Handler {
	return &Handler{
		groupDB:       groupDB,
		groupMemberDB: groupMemberDB,
		expenseDB:     expenseDB,
		expenseUserDB: expenseUserDB,
		userBalanceDB: userBalanceDB,
	}
}
