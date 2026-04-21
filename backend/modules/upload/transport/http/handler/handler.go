package handler

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	coreauth "goadmin/core/auth"
	coreauthjwt "goadmin/core/auth/jwt"
	apperrors "goadmin/core/errors"
	coremiddleware "goadmin/core/middleware"
	"goadmin/core/response"
	coretenant "goadmin/core/tenant"
	coretransport "goadmin/core/transport"
	uploadservice "goadmin/modules/upload/application/service"
	uploadmodel "goadmin/modules/upload/domain/model"
	uploadrepo "goadmin/modules/upload/domain/repository"
	uploadreq "goadmin/modules/upload/transport/http/request"
	uploadresp "goadmin/modules/upload/transport/http/response"

	"go.uber.org/zap"
)

type Handler struct {
	service *uploadservice.Service
	logger  *zap.Logger
}

func New(service *uploadservice.Service, logger *zap.Logger) *Handler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Handler{service: service, logger: logger}
}

func (h *Handler) List(c coretransport.Context) {
	var req uploadreq.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, body := response.Failure(apperrors.Wrap(err, apperrors.CodeBadRequest, "invalid list request"), requestID(c))
		c.JSON(status, body)
		return
	}
	items, total, err := h.service.List(c.RequestContext(), uploadrepo.ListFilter{
		Keyword:    req.Keyword,
		Visibility: req.Visibility,
		Status:     req.Status,
		BizModule:  req.BizModule,
		BizType:    req.BizType,
		BizId:      req.BizId,
		UploadedBy: req.UploadedBy,
		Page:       req.Page,
		PageSize:   req.PageSize,
	})
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(uploadresp.List{Total: total, Items: mapItems(items)}, requestID(c)))
}

func (h *Handler) Get(c coretransport.Context) {
	item, err := h.service.Get(c.RequestContext(), c.Param("id"))
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapItem(*item), requestID(c)))
}

func (h *Handler) Upload(c coretransport.Context) {
	header, err := c.FormFile("file")
	if err != nil {
		status, body := response.Failure(apperrors.Wrap(err, apperrors.CodeBadRequest, "upload file is required"), requestID(c))
		c.JSON(status, body)
		return
	}
	file, err := header.Open()
	if err != nil {
		status, body := response.Failure(apperrors.Wrap(err, apperrors.CodeInternal, "open upload file"), requestID(c))
		c.JSON(status, body)
		return
	}
	defer func() { _ = file.Close() }()

	var req uploadreq.UploadRequest
	if err := c.ShouldBind(&req); err != nil {
		status, body := response.Failure(apperrors.Wrap(err, apperrors.CodeBadRequest, "invalid upload request"), requestID(c))
		c.JSON(status, body)
		return
	}
	item, err := h.service.Upload(c.RequestContext(), uploadservice.UploadRequest{
		File:        file,
		Filename:    header.Filename,
		ContentType: header.Header.Get("Content-Type"),
		Size:        header.Size,
		Visibility:  req.Visibility,
		BizModule:   req.BizModule,
		BizType:     req.BizType,
		BizId:       req.BizId,
		BizField:    req.BizField,
		Remark:      req.Remark,
		UploadedBy:  requestUser(c),
		TenantId:    requestTenant(c),
	})
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusCreated, response.Success(mapItem(*item), requestID(c)))
}

func (h *Handler) Delete(c coretransport.Context) {
	if err := h.service.Delete(c.RequestContext(), c.Param("id")); err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(map[string]any{"deleted": true}, requestID(c)))
}

func (h *Handler) Download(c coretransport.Context) {
	reader, info, item, err := h.service.Open(c.RequestContext(), c.Param("id"))
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	defer func() { _ = reader.Close() }()

	ext := strings.TrimSpace(item.Extension)
	if ext == "" {
		ext = filepath.Ext(item.OriginalName)
	}
	tmp, err := os.CreateTemp("", "goadmin-upload-*"+ext)
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	if _, err := io.Copy(tmp, reader); err != nil {
		_ = tmp.Close()
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	if err := tmp.Close(); err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	if info != nil && strings.TrimSpace(info.ContentType) != "" {
		c.SetHeader("Content-Type", info.ContentType)
	}
	c.FileAttachment(tmpPath, item.OriginalName)
}

func (h *Handler) Preview(c coretransport.Context) {
	item, err := h.service.Get(c.RequestContext(), c.Param("id"))
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapPreview(*item), requestID(c)))
}

func (h *Handler) Bind(c coretransport.Context) {
	var req uploadreq.BindRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := response.Failure(apperrors.Wrap(err, apperrors.CodeBadRequest, "invalid bind request"), requestID(c))
		c.JSON(status, body)
		return
	}
	item, err := h.service.Bind(c.RequestContext(), c.Param("id"), uploadmodel.FileBinding{
		BizModule: req.BizModule,
		BizType:   req.BizType,
		BizId:     req.BizId,
		BizField:  req.BizField,
	})
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapItem(*item), requestID(c)))
}

func (h *Handler) Unbind(c coretransport.Context) {
	item, err := h.service.Unbind(c.RequestContext(), c.Param("id"))
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapItem(*item), requestID(c)))
}

func (h *Handler) GetDefaultStorage(c coretransport.Context) {
	driver, err := h.service.DefaultStorageDriver(c.RequestContext())
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(uploadresp.StorageSetting{Driver: driver}, requestID(c)))
}

func (h *Handler) SetDefaultStorage(c coretransport.Context) {
	var req uploadreq.StorageSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := response.Failure(apperrors.Wrap(err, apperrors.CodeBadRequest, "invalid storage setting request"), requestID(c))
		c.JSON(status, body)
		return
	}
	if err := h.service.SetDefaultStorageDriver(c.RequestContext(), req.Driver); err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(uploadresp.StorageSetting{Driver: strings.TrimSpace(req.Driver)}, requestID(c)))
}

func mapItems(items []uploadmodel.FileAsset) []uploadresp.FileItem {
	result := make([]uploadresp.FileItem, 0, len(items))
	for _, item := range items {
		result = append(result, mapItem(item))
	}
	return result
}

func mapItem(item uploadmodel.FileAsset) uploadresp.FileItem {
	return uploadresp.FileItem{
		Id:             item.Id,
		TenantId:       item.TenantId,
		OriginalName:   item.OriginalName,
		StorageName:    item.StorageName,
		StorageKey:     item.StorageKey,
		StorageDriver:  item.StorageDriver,
		StoragePath:    item.StoragePath,
		PublicURL:      item.PublicURL,
		MimeType:       item.MimeType,
		Extension:      item.Extension,
		SizeBytes:      item.SizeBytes,
		ChecksumSHA256: item.ChecksumSHA256,
		Visibility:     string(item.Visibility),
		BizModule:      item.BizModule,
		BizType:        item.BizType,
		BizId:          item.BizId,
		BizField:       item.BizField,
		UploadedBy:     item.UploadedBy,
		Status:         string(item.Status),
		Remark:         item.Remark,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
	}
}

func mapPreview(item uploadmodel.FileAsset) uploadresp.Preview {
	fileItem := mapItem(item)
	kind := resolvePreviewKind(item.MimeType)
	mode := "download_only"
	if item.Visibility == uploadmodel.FileVisibilityPublic && strings.TrimSpace(item.PublicURL) != "" {
		mode = "public_url"
	} else if kind != "download-only" {
		mode = "download"
	}
	return uploadresp.Preview{
		FileItem:         fileItem,
		PreviewKind:      kind,
		PreviewMode:      mode,
		DownloadURL:      buildDownloadURL(item.Id),
		CanPreview:       kind != "download-only",
		CanOpenInBrowser: kind != "download-only",
	}
}

func resolvePreviewKind(mimeType string) string {
	normalized := strings.ToLower(strings.TrimSpace(mimeType))
	if normalized == "" {
		return "download-only"
	}
	if strings.HasPrefix(normalized, "image/") {
		return "image"
	}
	if normalized == "application/pdf" {
		return "pdf"
	}
	if strings.HasPrefix(normalized, "text/") || normalized == "application/json" || normalized == "application/xml" || normalized == "application/xhtml+xml" {
		return "text"
	}
	return "download-only"
}

func buildDownloadURL(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return ""
	}
	return "/api/v1/uploads/files/" + id + "/download"
}

func requestID(c coretransport.Context) string {
	if value, exists := c.Get(coremiddleware.RequestIDContextKey); exists {
		if id, ok := value.(string); ok {
			return id
		}
	}
	return ""
}

func requestUser(c coretransport.Context) string {
	if value, exists := c.Get("auth.claims"); exists {
		if claims, ok := value.(*coreauthjwt.Claims); ok && claims != nil {
			return strings.TrimSpace(claims.Username)
		}
	}
	if claims, ok := coreauth.ClaimsFromContext(c.RequestContext()); ok {
		return strings.TrimSpace(claims.Username)
	}
	return ""
}

func requestTenant(c coretransport.Context) string {
	if tenant, ok := coretenant.TenantFromContext(c.RequestContext()); ok {
		return strings.TrimSpace(tenant.ID)
	}
	return ""
}
