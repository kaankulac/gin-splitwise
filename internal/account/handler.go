package account

import (
	accountDB "gin-splitwise/internal/account/database"
	"gin-splitwise/internal/account/model"
	"gin-splitwise/internal/config"
	"gin-splitwise/internal/database"
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
	accountDB accountDB.AccountDB
}

// POST /v1/api/users
func (h *Handler) signUp(c *gin.Context) {
	handler.HandleRequest(c, func(c *gin.Context) *handler.Response {
		logger := logging.FromContext(c)
		type RequestBody struct {
			User struct {
				Username string `json:"username" binding:"required"`
				Email string `json:"email" binding:"email"`
				Password string `json:"password" binding:"required,min=8"`
			} `json:"user"`
		}
		var body RequestBody
		if err := c.ShouldBindJSON(&body); err != nil {
			logger.Errorw("account.handler.signUp failed to bind", "err", err)
			var details []*validate.ValidationErrDetail
			if vErrs, ok := err.(validator.ValidationErrors); ok {
				details = validate.ValidationErrorDetails(&body.User, "json", vErrs)
			}
			return handler.NewErrorResponse(http.StatusBadRequest, handler.InvalidBodyValue, "invalid user payload in body", details)
		}

		password, err := EncodePassword(body.User.Password)
		if err != nil {
			logger.Errorw("account.handler.signUp failed to encode password", "err", err)
			return handler.NewInternalServerErrorResponse(err)
		}
		acc := model.Account{
			Username: body.User.Username,
			Email: body.User.Email,
			Password: password,
		}
		err = h.accountDB.Save(c.Request.Context(), &acc)
		if err != nil {
			if database.IsKeyConflictError(err) {
				return handler.NewErrorResponse(http.StatusConflict, handler.DuplicateEntity, "account with this email already exists", nil)
			}
			return handler.NewInternalServerErrorResponse(err)
		}
		return handler.NewSuccessResponse(http.StatusCreated, NewUserResponse(&acc))
	})
}

// GET /v1/api/user/me
func (h *Handler) currentUser(c *gin.Context) {
	handler.HandleRequest(c, func(c *gin.Context) *handler.Response {
		currentUser := shared.MustCurrentUser(c)
		find, err := h.accountDB.FindByEmail(c.Request.Context(), currentUser.Email)
		if err != nil || find.Disabled {
			if database.IsRecordNotFoundErr(err) || find.Disabled {
				return handler.NewErrorResponse(http.StatusNotFound, handler.NotFoundEntity, "current user not found", nil)
			}
			return &handler.Response{Err: err}
		}
		return handler.NewSuccessResponse(http.StatusOK, NewUserResponse(find))
	})
}

// PUT /v1/api/user
func (h *Handler) update(c *gin.Context) {
	handler.HandleRequest(c, func(c *gin.Context) *handler.Response {
		logger := logging.FromContext(c)
		currentUser := shared.MustCurrentUser(c)
		type RequestBody struct {
			User struct {
				Username string `json:"username" binding:"omitempty"`
				Password string `json:"password" binding:"omitempty,min=8"`
				Image string `json:"image" binding:"omitempty"`
			} `json:"user"`
		}
		var body RequestBody
		if err := c.ShouldBindJSON(&body); err != nil {
			logger.Errorw("account.handler.update failed to bind", "err", err)
			var details []*validate.ValidationErrDetail
			if vErrs, ok := err.(validator.ValidationErrors); ok {
				details = validate.ValidationErrorDetails(&body.User, "json", vErrs)
			}
			return handler.NewErrorResponse(http.StatusBadRequest, handler.InvalidBodyValue, "invalid user payload in body", details)
		}

		acc, err := h.accountDB.FindByEmail(c.Request.Context(), currentUser.Email)
		if err != nil {
			if database.IsRecordNotFoundErr(err) {
				return handler.NewErrorResponse(http.StatusNotFound, handler.NotFoundEntity, "current user not found", nil)
			}
			return handler.NewInternalServerErrorResponse(err)
		}

		if body.User.Password != "" {
			password, err := EncodePassword(body.User.Password)
			if err != nil {
				logger.Errorw("account.handler.update failed to encode password", "err", err)
				return handler.NewInternalServerErrorResponse(err)
			}
			acc.Password = password
		}
		if body.User.Username != "" {
			acc.Username = body.User.Username
		}
		if body.User.Image != "" {
			acc.Image = body.User.Image
		}
		err = h.accountDB.Update(c.Request.Context(), currentUser.Email, acc)
		if err != nil {
			if database.IsRecordNotFoundErr(err) {
				logger.Errorw("account.handler.update failed to update user because not found user", "err", err)
			}
			return handler.NewInternalServerErrorResponse(err)
		}
		return handler.NewSuccessResponse(http.StatusOK, NewUserResponse(acc))
	})
}

func RouteV1(cfg *config.Config, h *Handler, r *gin.Engine, auth *jwt.GinJWTMiddleware) {
	v1 := r.Group("v1/api")
	v1.Use(middleware.RequestIDMiddleware(), middleware.TimeoutMiddleware(cfg.ServerConfig.WriteTimeout))

	// Non-Authenticated Routes
	v1.Use()
	{
		v1.POST("users/login", auth.LoginHandler)
		v1.POST("users/signup", h.signUp)
	}
	
	// Authenticated Routes
	v1.Use(auth.MiddlewareFunc())
	{
		v1.GET("/users/profile", h.currentUser)
		v1.PUT("/users/profile", h.update)
		v1.POST("/groups")
	}
}

func NewHandler(accountDB accountDB.AccountDB) *Handler {
	return &Handler{
		accountDB: accountDB,
	}
}