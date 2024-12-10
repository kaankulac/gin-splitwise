package group

import (
	"gin-splitwise/internal/group/model"
	"time"
)

type GroupResponse struct {
	Group Group `json:"group"`
}

type GroupsResponse struct {
	Groups []Group `json:"groups"`
}

type Group struct {
	ID uint `json:"id"`
	Title string `json:"title"`
	Description string `json:"description"`
	Image string `json:"image"`
	OwnerId uint `json:"owner_id"`
	TotalExpense uint `json:"total_expense"`
	TotalMember uint `json:"total_member"`
	Members []Member `json:"members"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Member struct {
	ID uint `json:"id"`
	Username string `json:"name"`
}

func NewGroupResponse(group *model.Group) *GroupResponse {
	return &GroupResponse{
		Group: Group{
			ID: group.ID,
			Title: group.Title,
			Description: group.Description,
			Image: group.Image,
			OwnerId: group.OwnerId,
			TotalExpense: group.TotalExpense,
			TotalMember: group.TotalMember,
			CreatedAt: group.CreatedAt,
			UpdatedAt: group.UpdatedAt,
		},
	}
}

func NewGroupWithMembersResponse(group *model.Group, members []Member) *GroupResponse {
	return &GroupResponse{
		Group: Group{
			ID: group.ID,
			Title: group.Title,
			Description: group.Description,
			Image: group.Image,
			OwnerId: group.OwnerId,
			TotalExpense: group.TotalExpense,
			Members: members,
			TotalMember: group.TotalMember,
			CreatedAt: group.CreatedAt,
			UpdatedAt: group.UpdatedAt,
		},
	}
}

func NewGroupsResponse(groups []*model.Group) *GroupsResponse {
	var groupResponse []Group
	for _, group := range groups {
		groupResponse = append(groupResponse, Group{
			ID: group.ID,
			Title: group.Title,
			Description: group.Description,
			Image: group.Image,
			OwnerId: group.OwnerId,
			TotalExpense: group.TotalExpense,
			TotalMember: group.TotalMember,
			CreatedAt: group.CreatedAt,
			UpdatedAt: group.UpdatedAt,
		})
	}

	return &GroupsResponse{
		Groups: groupResponse,
	}
}