package handler

import (
	"net/http"
	"sort"
	"strings"

	apperrors "goadmin/core/errors"
	coremiddleware "goadmin/core/middleware"
	"goadmin/core/response"
	coretransport "goadmin/core/transport"
	plugincommand "goadmin/plugin/application/command"
	pluginservice "goadmin/plugin/application/service"
	pluginmodel "goadmin/plugin/domain/model"
	pluginiface "goadmin/plugin/interface"
	pluginreq "goadmin/plugin/transport/http/request"
	pluginresp "goadmin/plugin/transport/http/response"

	"go.uber.org/zap"
)

type Handler struct {
	service *pluginservice.Service
	logger  *zap.Logger
}

func New(service *pluginservice.Service, logger *zap.Logger) *Handler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Handler{service: service, logger: logger}
}

func (h *Handler) List(c coretransport.Context) {
	items, total, err := h.service.List(c.RequestContext())
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(pluginresp.List{Total: total, Items: mapPlugins(items)}, requestID(c)))
}

func (h *Handler) Get(c coretransport.Context) {
	item, err := h.service.Get(c.RequestContext(), c.Param("name"))
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapPlugin(*item), requestID(c)))
}

func (h *Handler) Create(c coretransport.Context) {
	var req pluginreq.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := response.Failure(apperrors.Wrap(err, apperrors.CodeBadRequest, "invalid create request"), requestID(c))
		c.JSON(status, body)
		return
	}
	item, err := h.service.Create(c.RequestContext(), plugincommand.CreatePlugin{
		Name:        req.Name,
		Description: req.Description,
		Enabled:     req.Enabled,
		Menus:       mapMenus(req.Name, req.Menus),
		Permissions: mapPermissions(req.Name, req.Permissions),
	})
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusCreated, response.Success(mapPlugin(*item), requestID(c)))
}

func (h *Handler) Update(c coretransport.Context) {
	var req pluginreq.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := response.Failure(apperrors.Wrap(err, apperrors.CodeBadRequest, "invalid update request"), requestID(c))
		c.JSON(status, body)
		return
	}
	var menus []pluginiface.Menu
	if req.Menus != nil {
		menus = mapMenus(c.Param("name"), req.Menus)
	}
	var permissions []pluginiface.Permission
	if req.Permissions != nil {
		permissions = mapPermissions(c.Param("name"), req.Permissions)
	}
	item, err := h.service.Update(c.RequestContext(), c.Param("name"), plugincommand.UpdatePlugin{
		Description: req.Description,
		Enabled:     req.Enabled,
		Menus:       menus,
		Permissions: permissions,
	})
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapPlugin(*item), requestID(c)))
}

func (h *Handler) Delete(c coretransport.Context) {
	if err := h.service.Delete(c.RequestContext(), c.Param("name")); err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(map[string]any{"deleted": true}, requestID(c)))
}

func (h *Handler) Menus(c coretransport.Context) {
	items, err := h.service.Menus(c.RequestContext())
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(pluginresp.MenuList{Items: buildMenuTree(items)}, requestID(c)))
}

func (h *Handler) Permissions(c coretransport.Context) {
	items, err := h.service.Permissions(c.RequestContext())
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(pluginresp.PermissionList{Items: mapPermissionsResponse(items)}, requestID(c)))
}

func requestID(c coretransport.Context) string {
	if value, exists := c.Get(coremiddleware.RequestIDContextKey); exists {
		if id, ok := value.(string); ok {
			return id
		}
	}
	return ""
}

func mapPlugins(items []pluginmodel.Plugin) []pluginresp.Item {
	result := make([]pluginresp.Item, 0, len(items))
	for _, item := range items {
		result = append(result, mapPlugin(item))
	}
	return result
}

func mapPlugin(item pluginmodel.Plugin) pluginresp.Item {
	return pluginresp.Item{
		Name:        item.Name,
		Description: item.Description,
		Enabled:     item.Enabled,
		Menus:       mapMenusResponse(item.Menus),
		Permissions: mapPermissionsResponse(item.Permissions),
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func mapMenus(pluginName string, items []pluginreq.Menu) []pluginiface.Menu {
	if items == nil {
		return nil
	}
	result := make([]pluginiface.Menu, 0, len(items))
	for _, item := range items {
		result = append(result, pluginiface.Menu{
			Plugin:      fallbackPlugin(pluginName, item.Plugin),
			ID:          strings.TrimSpace(item.ID),
			ParentID:    strings.TrimSpace(item.ParentID),
			Name:        strings.TrimSpace(item.Name),
			Path:        strings.TrimSpace(item.Path),
			Component:   strings.TrimSpace(item.Component),
			Icon:        strings.TrimSpace(item.Icon),
			Sort:        item.Sort,
			Permission:  strings.TrimSpace(item.Permission),
			Type:        pluginiface.MenuType(strings.TrimSpace(item.Type)),
			Visible:     item.Visible,
			Enabled:     item.Enabled,
			Redirect:    strings.TrimSpace(item.Redirect),
			ExternalURL: strings.TrimSpace(item.ExternalURL),
		})
	}
	return result
}

func mapPermissions(pluginName string, items []pluginreq.Permission) []pluginiface.Permission {
	if items == nil {
		return nil
	}
	result := make([]pluginiface.Permission, 0, len(items))
	for _, item := range items {
		result = append(result, pluginiface.Permission{
			Plugin:      pluginName,
			Object:      strings.TrimSpace(item.Object),
			Action:      strings.TrimSpace(item.Action),
			Description: strings.TrimSpace(item.Description),
		})
	}
	return result
}

func mapMenusResponse(items []pluginiface.Menu) []pluginresp.Menu {
	result := make([]pluginresp.Menu, 0, len(items))
	for _, item := range items {
		result = append(result, pluginresp.Menu{
			Plugin:      item.Plugin,
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
			Children:    mapMenusResponse(item.Children),
		})
	}
	return result
}

func mapPermissionsResponse(items []pluginiface.Permission) []pluginresp.Permission {
	result := make([]pluginresp.Permission, 0, len(items))
	for _, item := range items {
		result = append(result, pluginresp.Permission{
			Plugin:      item.Plugin,
			Object:      item.Object,
			Action:      item.Action,
			Description: item.Description,
		})
	}
	return result
}

func buildMenuTree(items []pluginiface.Menu) []pluginresp.Menu {
	if len(items) == 0 {
		return nil
	}
	byParent := make(map[string][]pluginiface.Menu)
	for _, item := range items {
		byParent[menuParentKey(item)] = append(byParent[menuParentKey(item)], item)
	}
	for key := range byParent {
		sort.Slice(byParent[key], func(i, j int) bool {
			if byParent[key][i].Sort == byParent[key][j].Sort {
				if byParent[key][i].Plugin == byParent[key][j].Plugin {
					return byParent[key][i].Name < byParent[key][j].Name
				}
				return byParent[key][i].Plugin < byParent[key][j].Plugin
			}
			return byParent[key][i].Sort < byParent[key][j].Sort
		})
	}
	roots := byParent[menuRootKey("")]
	for _, item := range items {
		if strings.TrimSpace(item.ParentID) == "" && len(roots) == 0 {
			roots = append(roots, item)
		}
	}
	if len(roots) == 0 {
		roots = make([]pluginiface.Menu, 0)
		for _, item := range items {
			if strings.TrimSpace(item.ParentID) == "" {
				roots = append(roots, item)
			}
		}
	}
	return mapMenuChildren(roots, byParent)
}

func mapMenuChildren(items []pluginiface.Menu, byParent map[string][]pluginiface.Menu) []pluginresp.Menu {
	result := make([]pluginresp.Menu, 0, len(items))
	for _, item := range items {
		result = append(result, pluginresp.Menu{
			Plugin:      item.Plugin,
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
			Children:    mapMenuChildren(byParent[menuNodeKey(item)], byParent),
		})
	}
	return result
}

func menuRootKey(plugin string) string {
	return strings.ToLower(strings.TrimSpace(plugin)) + ":root"
}

func menuParentKey(item pluginiface.Menu) string {
	plugin := strings.ToLower(strings.TrimSpace(item.Plugin))
	parentID := strings.TrimSpace(item.ParentID)
	if parentID == "" {
		return plugin + ":root"
	}
	return plugin + ":id:" + strings.ToLower(parentID)
}

func menuNodeKey(item pluginiface.Menu) string {
	plugin := strings.ToLower(strings.TrimSpace(item.Plugin))
	if id := strings.TrimSpace(item.ID); id != "" {
		return plugin + ":id:" + strings.ToLower(id)
	}
	return plugin + ":path:" + strings.ToLower(strings.TrimSpace(item.Path))
}

func fallbackPlugin(primary, secondary string) string {
	if strings.TrimSpace(primary) != "" {
		return strings.TrimSpace(primary)
	}
	return strings.TrimSpace(secondary)
}
