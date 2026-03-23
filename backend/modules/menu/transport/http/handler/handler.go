package handler

import (
	"net/http"
	"strings"

	apperrors "goadmin/core/errors"
	coremiddleware "goadmin/core/middleware"
	"goadmin/core/response"
	coretransport "goadmin/core/transport"
	"goadmin/modules/menu/application/command"
	"goadmin/modules/menu/application/query"
	menuservice "goadmin/modules/menu/application/service"
	"goadmin/modules/menu/domain/model"
	menureq "goadmin/modules/menu/transport/http/request"
	menuresp "goadmin/modules/menu/transport/http/response"

	"go.uber.org/zap"
)

type Handler struct {
	service *menuservice.Service
	logger  *zap.Logger
}

func New(service *menuservice.Service, logger *zap.Logger) *Handler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Handler{service: service, logger: logger}
}

func (h *Handler) List(c coretransport.Context) {
	var req menureq.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, body := response.Failure(apperrors.Wrap(err, apperrors.CodeBadRequest, "invalid list request"), requestID(c))
		c.JSON(status, body)
		return
	}
	items, total, err := h.service.List(c.RequestContext(), query.ListMenus{
		Keyword:  req.Keyword,
		ParentID: req.ParentID,
		Visible:  req.Visible,
		Enabled:  req.Enabled,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(menuresp.List{Total: total, Items: mapMenus(items)}, requestID(c)))
}

func (h *Handler) Get(c coretransport.Context) {
	item, err := h.service.Get(c.RequestContext(), c.Param("id"))
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapMenu(*item), requestID(c)))
}

func (h *Handler) Tree(c coretransport.Context) {
	items, err := h.service.Tree(c.RequestContext())
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(menuresp.Tree{Items: mapMenus(items)}, requestID(c)))
}

func (h *Handler) Routes(c coretransport.Context) {
	items, err := h.service.Tree(c.RequestContext())
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(menuresp.Routes{Items: mapRoutes(items)}, requestID(c)))
}

func (h *Handler) Create(c coretransport.Context) {
	var req menureq.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := response.Failure(apperrors.Wrap(err, apperrors.CodeBadRequest, "invalid create request"), requestID(c))
		c.JSON(status, body)
		return
	}
	item, err := h.service.Create(c.RequestContext(), command.CreateMenu{
		ParentID:    req.ParentID,
		Name:        req.Name,
		Path:        req.Path,
		Component:   req.Component,
		Icon:        req.Icon,
		Sort:        req.Sort,
		Permission:  req.Permission,
		Type:        req.Type,
		Visible:     req.Visible,
		Enabled:     req.Enabled,
		Redirect:    req.Redirect,
		ExternalURL: req.ExternalURL,
	})
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusCreated, response.Success(mapMenu(*item), requestID(c)))
}

func (h *Handler) Update(c coretransport.Context) {
	var req menureq.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := response.Failure(apperrors.Wrap(err, apperrors.CodeBadRequest, "invalid update request"), requestID(c))
		c.JSON(status, body)
		return
	}
	item, err := h.service.Update(c.RequestContext(), c.Param("id"), command.UpdateMenu{
		ParentID:    req.ParentID,
		Name:        req.Name,
		Path:        req.Path,
		Component:   req.Component,
		Icon:        req.Icon,
		Sort:        req.Sort,
		Permission:  req.Permission,
		Type:        req.Type,
		Visible:     req.Visible,
		Enabled:     req.Enabled,
		Redirect:    req.Redirect,
		ExternalURL: req.ExternalURL,
	})
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapMenu(*item), requestID(c)))
}

func (h *Handler) Delete(c coretransport.Context) {
	if err := h.service.Delete(c.RequestContext(), c.Param("id")); err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(map[string]any{"deleted": true}, requestID(c)))
}

func mapMenus(items []model.Menu) []menuresp.Item {
	result := make([]menuresp.Item, 0, len(items))
	for _, item := range items {
		result = append(result, mapMenu(item))
	}
	return result
}

func mapMenu(item model.Menu) menuresp.Item {
	children := make([]menuresp.Item, 0, len(item.Children))
	for _, child := range item.Children {
		children = append(children, mapMenu(child))
	}
	return menuresp.Item{
		ID:          item.ID,
		ParentID:    item.ParentID,
		Name:        item.Name,
		Path:        item.Path,
		Component:   item.Component,
		Icon:        item.Icon,
		Sort:        item.Sort,
		Permission:  item.Permission,
		Type:        string(item.Type),
		Visible:     item.Visible,
		Enabled:     item.Enabled,
		Redirect:    item.Redirect,
		ExternalURL: item.ExternalURL,
		Children:    children,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func mapRoutes(items []model.Menu) []menuresp.Route {
	result := make([]menuresp.Route, 0, len(items))
	for _, item := range items {
		if !item.Enabled {
			continue
		}
		children := mapRoutes(item.Children)
		alwaysShow := len(children) > 0 && item.Type == model.TypeDirectory
		result = append(result, menuresp.Route{
			Name:       routeName(item),
			Path:       item.Path,
			Component:  item.Component,
			Redirect:   item.Redirect,
			Hidden:     !item.Visible,
			AlwaysShow: alwaysShow,
			Meta: menuresp.RouteMeta{
				Title:      item.Name,
				Icon:       item.Icon,
				Permission: item.Permission,
				Hidden:     !item.Visible,
				NoCache:    item.Type == model.TypeButton,
				Affix:      item.Path == "/dashboard",
				Link:       item.ExternalURL,
			},
			Children: children,
		})
	}
	return result
}

func routeName(item model.Menu) string {
	if name := routeNameFromPath(item.Path); name != "" {
		return name
	}
	if name := routeNameFromText(item.Name); name != "" {
		return name
	}
	if name := routeNameFromText(item.Permission); name != "" {
		return name
	}
	return "route"
}

func routeNameFromPath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" || trimmed == "/" {
		return ""
	}
	parts := strings.Split(strings.Trim(trimmed, "/"), "/")
	return camelizeParts(parts)
}

func routeNameFromText(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	parts := strings.FieldsFunc(trimmed, func(r rune) bool {
		return r == ' ' || r == '-' || r == '_' || r == ':' || r == '/'
	})
	return camelizeParts(parts)
}

func camelizeParts(parts []string) string {
	result := strings.Builder{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if result.Len() == 0 {
			result.WriteString(strings.ToLower(part[:1]))
			if len(part) > 1 {
				result.WriteString(strings.ToLower(part[1:]))
			}
			continue
		}
		result.WriteString(strings.ToUpper(part[:1]))
		if len(part) > 1 {
			result.WriteString(strings.ToLower(part[1:]))
		}
	}
	return result.String()
}

func requestID(c coretransport.Context) string {
	if value, exists := c.Get(coremiddleware.RequestIDContextKey); exists {
		if id, ok := value.(string); ok {
			return id
		}
	}
	return ""
}
