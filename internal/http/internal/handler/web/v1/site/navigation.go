package site

import (
	"go-chat/internal/pkg/ichat"
	"go-chat/internal/service"
)

type Navigation struct {
	service *service.NavigationService
}

func NewNavigation(service *service.NavigationService) *Navigation {
	return &Navigation{service: service}
}

func (c *Navigation) List(ctx *ichat.Context) error {
	items, err := c.service.FindList()
	if err != nil {
		return ctx.BusinessError(err.Error())
	}
	return ctx.Success(items)
}
