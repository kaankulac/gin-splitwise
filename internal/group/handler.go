package group

import (
	accountDB "gin-splitwise/internal/account/database"
	"gin-splitwise/internal/config"
	"gin-splitwise/internal/database"
	groupDB "gin-splitwise/internal/group/database"
	"gin-splitwise/internal/group/model"
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
	groupDB       groupDB.GroupDB
	groupMemberDB groupDB.GroupMemberDB
	accountDB     accountDB.AccountDB
}

// POST /v1/api/groups
func (h *Handler) createGroup(c *gin.Context) {
	handler.HandleRequest(c, func(c *gin.Context) *handler.Response {
		logger := logging.FromContext(c)
		type RequestBody struct {
			Group struct {
				Title       string `json:"title"`
				Description string `json:"description"`
				Image       string `json:"image"`
			} `json:"group"`
		}
		var body RequestBody
		if err := c.ShouldBindJSON(&body); err != nil {
			logger.Errorw("group.handler.createGroup failed to bind", "err", err)
			var details []*validate.ValidationErrDetail
			if vErrs, ok := err.(validator.ValidationErrors); ok {
				details = validate.ValidationErrorDetails(&body.Group, "json", vErrs)
			}
			return handler.NewErrorResponse(http.StatusBadRequest, handler.InvalidBodyValue, "invalid group payload in body", details)
		}

		currentUser := shared.MustCurrentUser(c)

		group := model.Group{
			Title:        body.Group.Title,
			Description:  body.Group.Description,
			Image:        body.Group.Image,
			TotalExpense: 0,
			TotalMember:  0,
			OwnerId:      currentUser.ID,
		}

		err := h.groupDB.Save(c.Request.Context(), &group)
		if err != nil {
			return handler.NewInternalServerErrorResponse(err)
		}
		return handler.NewSuccessResponse(http.StatusCreated, NewGroupResponse(&group))
	})
}

func (h *Handler) updateGroup(c *gin.Context) {
	handler.HandleRequest(c, func(c *gin.Context) *handler.Response {
		logger := logging.FromContext(c)
		currentUser := shared.MustCurrentUser(c)
		type RequestUri struct {
			GroupId uint `uri:"groupId" binding:"required,numeric"`
		}
		type RequestBody struct {
			Group struct {
				Title       string `json:"title" binding:"omitempty"`
				Description string `json:"description" binding:"omitempty"`
				Image       string `json:"image" binding:"omitempty"`
			} `json:"group"`
		}
		var uri RequestUri
		var body RequestBody
		if err := c.ShouldBindUri(&uri); err != nil {
			logger.Errorw("group.handler.updateGroup failed to bind uri", "err", err)
			var details []*validate.ValidationErrDetail
			if vErrs, ok := err.(validator.ValidationErrors); ok {
				details = validate.ValidationErrorDetails(&uri, "uri", vErrs)
			}
			return handler.NewErrorResponse(http.StatusBadRequest, handler.InvalidUriValue, "invalid group id in uri", details)
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			logger.Errorw("group.handler.updateGroup failed to bind body", "err", err)
			var details []*validate.ValidationErrDetail
			if vErrs, ok := err.(validator.ValidationErrors); ok {
				details = validate.ValidationErrorDetails(&body.Group, "json", vErrs)
			}
			return handler.NewErrorResponse(http.StatusBadRequest, handler.InvalidBodyValue, "invalid group payload in body", details)
		}

		if isOwner, err := h.groupDB.CheckOwner(c.Request.Context(), uri.GroupId, currentUser.ID); !isOwner || err != nil {
			if err != nil {
				return handler.NewInternalServerErrorResponse(err)
			}
			return handler.NewErrorResponse(http.StatusForbidden, handler.Forbidden, "current user is not owner of the group", nil)
		}

		group := model.Group{
			Title:       body.Group.Title,
			Description: body.Group.Description,
			Image:       body.Group.Image,
		}

		err := h.groupDB.Update(c.Request.Context(), uri.GroupId, &group)
		if err != nil {
			return handler.NewInternalServerErrorResponse(err)
		}
		return handler.NewSuccessResponse(http.StatusOK, NewGroupResponse(&group))
	})
}

func (h *Handler) addGroupMember(c *gin.Context) {
	handler.HandleRequest(c, func(c *gin.Context) *handler.Response {
		logger := logging.FromContext(c)
		currentUser := shared.MustCurrentUser(c)
		type RequestBody struct {
			GroupMember struct {
				Email   string `json:"email" binding:"required"`
				GroupId uint   `json:"groupId" binding:"required,numeric"`
			}
		}
		var body RequestBody
		if err := c.ShouldBindJSON(&body); err != nil {
			logger.Errorw("group.handler.addGroupMember failed to bind", "err", err)
			var details []*validate.ValidationErrDetail
			if vErrs, ok := err.(validator.ValidationErrors); ok {
				details = validate.ValidationErrorDetails(&body.GroupMember, "json", vErrs)
			}
			return handler.NewErrorResponse(http.StatusBadRequest, handler.InvalidBodyValue, "invalid group member payload in body", details)
		}

		if isOwner, err := h.groupDB.CheckOwner(c.Request.Context(), body.GroupMember.GroupId, currentUser.ID); !isOwner || err != nil {
			if err != nil {
				return handler.NewInternalServerErrorResponse(err)
			}
			return handler.NewErrorResponse(http.StatusForbidden, handler.Forbidden, "current user is not owner of the group", nil)
		}

		acc, err := h.accountDB.FindByEmail(c.Request.Context(), body.GroupMember.Email)
		if err != nil {
			if database.IsRecordNotFoundErr(err) {
				return handler.NewErrorResponse(http.StatusNotFound, handler.NotFoundEntity, "user not found", nil)
			}
			return handler.NewInternalServerErrorResponse(err)
		}

		err = h.groupMemberDB.Add(c.Request.Context(), &model.GroupMember{
			GroupId:     body.GroupMember.GroupId,
			UserId:      acc.ID,
			TotalDebt:   0,
			TotalCredit: 0,
		})
		if err != nil {
			return handler.NewInternalServerErrorResponse(err)
		}
		h.groupDB.IncrementMemberCount(c.Request.Context(), body.GroupMember.GroupId)

		return handler.NewSuccessResponse(http.StatusOK, nil)
	})
}

func (h *Handler) leaveGroup(c *gin.Context) {
	handler.HandleRequest(c, func(c *gin.Context) *handler.Response {
		logger := logging.FromContext(c)
		currentUser := shared.MustCurrentUser(c)
		type RequestUri struct {
			GroupId uint `uri:"groupId" binding:"required,numeric"`
		}
		var uri RequestUri
		if err := c.ShouldBindUri(&uri); err != nil {
			logger.Errorw("group.handler.leaveGroup failed to bind uri", "err", err)
			var details []*validate.ValidationErrDetail
			if vErrs, ok := err.(validator.ValidationErrors); ok {
				details = validate.ValidationErrorDetails(&uri, "uri", vErrs)
			}
			return handler.NewErrorResponse(http.StatusBadRequest, handler.InvalidBodyValue, "invalid group id in uri", details)
		}

		if isOwner, err := h.groupDB.CheckOwner(c.Request.Context(), uri.GroupId, currentUser.ID); isOwner || err != nil {
			if err != nil {
				return handler.NewInternalServerErrorResponse(err)
			}
			return handler.NewErrorResponse(http.StatusForbidden, handler.Forbidden, "owner cannot leave the group", nil)
		}

		err := h.groupMemberDB.Remove(c.Request.Context(), uri.GroupId, currentUser.ID)
		if err != nil {
			return handler.NewInternalServerErrorResponse(err)
		}
		h.groupDB.DecrementMemberCount(c.Request.Context(), uri.GroupId)
		return handler.NewSuccessResponse(http.StatusOK, nil)
	})
}

func (h *Handler) deleteGroup(c *gin.Context) {
	handler.HandleRequest(c, func(c *gin.Context) *handler.Response {
		logger := logging.FromContext(c)
		currentUser := shared.MustCurrentUser(c)
		type RequestUri struct {
			GroupId uint `uri:"groupId" binding:"required,numeric"`
		}
		var uri RequestUri
		if err := c.ShouldBindUri(&uri); err != nil {
			logger.Errorw("group.handler.deleteGroup failed to bind uri", "err", err)
			var details []*validate.ValidationErrDetail
			if vErrs, ok := err.(validator.ValidationErrors); ok {
				details = validate.ValidationErrorDetails(&uri, "uri", vErrs)
			}
			return handler.NewErrorResponse(http.StatusBadRequest, handler.InvalidBodyValue, "invalid group id in uri", details)
		}

		if isOwner, err := h.groupDB.CheckOwner(c.Request.Context(), uri.GroupId, currentUser.ID); !isOwner || err != nil {
			if err != nil {
				return handler.NewInternalServerErrorResponse(err)
			}
			return handler.NewErrorResponse(http.StatusForbidden, handler.Forbidden, "current user is not owner of the group", nil)
		}
		err := h.groupDB.Delete(c.Request.Context(), uri.GroupId)
		if err != nil {
			return handler.NewInternalServerErrorResponse(err)
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
		v1.POST("/groups", h.createGroup)
		v1.PUT("/groups/:groupId", h.updateGroup)
		v1.DELETE("/groups/:groupId", h.deleteGroup)
		v1.POST("/groups/members", h.addGroupMember)
		v1.POST("/groups/:groupId/leave", h.leaveGroup)
	}
}

func NewHandler(groupDB groupDB.GroupDB, groupMemberDB groupDB.GroupMemberDB, accountDB accountDB.AccountDB) *Handler {
	return &Handler{
		groupDB:       groupDB,
		groupMemberDB: groupMemberDB,
		accountDB:     accountDB,
	}
}
