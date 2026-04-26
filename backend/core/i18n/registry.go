package i18n

import (
	"context"
	"sync"
)

type Registry struct {
	mu              sync.RWMutex
	resources       map[string]map[string]string
	defaultLanguage string
}

var defaultRegistry = newDefaultRegistry()

func NewRegistry(defaultLanguage string) *Registry {
	return &Registry{
		resources:       make(map[string]map[string]string),
		defaultLanguage: ResolveLanguage(defaultLanguage),
	}
}

func DefaultRegistry() *Registry {
	return defaultRegistry
}

func newDefaultRegistry() *Registry {
	registry := NewRegistry(DefaultLanguage())
	registry.Register(LanguageZHCN, map[string]string{
		"common.ok":                                   "成功",
		"common.created":                              "创建成功",
		"common.deleted":                              "删除成功",
		"common.logged_out":                           "已退出登录",
		"common.invalid_request":                      "请求参数错误",
		"common.permission_denied":                    "无权访问",
		"common.authentication_required":              "需要登录",
		"common.internal_error":                       "服务器内部错误",
		"common.route_not_found":                      "请求路由不存在",
		"auth.invalid_credentials":                    "用户名或密码错误",
		"auth.authentication_required":                "需要登录",
		"codegen.download.service_required":           "下载服务未配置",
		"codegen.download.dsl_required":               "需要 DSL 内容",
		"codegen.download.database_required":          "需要数据库连接",
		"codegen.download.task_id_required":           "任务 ID 不能为空",
		"codegen.download.artifact_expired":           "生成产物已过期",
		"codegen.download.artifact_not_found":         "未找到生成产物",
		"codegen.download.load_artifact_failed":       "加载生成产物失败",
		"codegen.download.cleanup_expired_failed":     "清理过期产物失败",
		"codegen.download.create_task_id_failed":      "创建下载任务失败",
		"codegen.download.create_workspace_failed":    "创建代码生成工作区失败",
		"codegen.download.collect_files_failed":       "收集生成文件失败",
		"codegen.download.prepare_package_failed":     "准备下载包失败",
		"codegen.download.create_package_file_failed": "创建临时包文件失败",
		"codegen.download.close_package_file_failed":  "关闭临时包文件失败",
		"codegen.download.package_artifact_failed":    "打包生成产物失败",
		"codegen.download.store_artifact_failed":      "保存生成产物失败",
	})
	registry.Register(LanguageENUS, map[string]string{
		"common.ok":                                   "ok",
		"common.created":                              "created successfully",
		"common.deleted":                              "deleted successfully",
		"common.logged_out":                           "logged out",
		"common.invalid_request":                      "invalid request",
		"common.permission_denied":                    "permission denied",
		"common.authentication_required":              "authentication required",
		"common.internal_error":                       "internal server error",
		"common.route_not_found":                      "route not found",
		"auth.invalid_credentials":                    "invalid username or password",
		"auth.authentication_required":                "authentication required",
		"codegen.download.service_required":           "download service is not configured",
		"codegen.download.dsl_required":               "DSL content is required",
		"codegen.download.database_required":          "database connection is required",
		"codegen.download.task_id_required":           "task id is required",
		"codegen.download.artifact_expired":           "generated artifact expired",
		"codegen.download.artifact_not_found":         "generated artifact not found",
		"codegen.download.load_artifact_failed":       "failed to load generated artifact",
		"codegen.download.cleanup_expired_failed":     "failed to clean up expired artifacts",
		"codegen.download.create_task_id_failed":      "failed to create download task id",
		"codegen.download.create_workspace_failed":    "failed to create codegen workspace",
		"codegen.download.collect_files_failed":       "failed to collect generated files",
		"codegen.download.prepare_package_failed":     "failed to prepare download package",
		"codegen.download.create_package_file_failed": "failed to create temporary package file",
		"codegen.download.close_package_file_failed":  "failed to close temporary package file",
		"codegen.download.package_artifact_failed":    "failed to package generated artifact",
		"codegen.download.store_artifact_failed":      "failed to store generated artifact",
	})
	return registry
}

func (r *Registry) Register(language string, entries map[string]string) {
	if r == nil || len(entries) == 0 {
		return
	}
	language = ResolveLanguage(language)
	if language == "" {
		language = r.defaultLanguage
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	bucket, ok := r.resources[language]
	if !ok {
		bucket = make(map[string]string, len(entries))
		r.resources[language] = bucket
	}
	for key, value := range entries {
		if key == "" || value == "" {
			continue
		}
		bucket[key] = value
	}
}

func (r *Registry) Translate(ctx context.Context, key string) (string, bool) {
	if r == nil {
		return "", false
	}
	return r.TranslateLanguage(LanguageOrDefault(ctx), key)
}

func (r *Registry) TranslateLanguage(language, key string) (string, bool) {
	if r == nil || key == "" {
		return "", false
	}
	language = ResolveLanguage(language)
	r.mu.RLock()
	defer r.mu.RUnlock()
	if bucket, ok := r.resources[language]; ok {
		if value, exists := bucket[key]; exists && value != "" {
			return value, true
		}
	}
	if language != r.defaultLanguage {
		if bucket, ok := r.resources[r.defaultLanguage]; ok {
			if value, exists := bucket[key]; exists && value != "" {
				return value, true
			}
		}
	}
	return "", false
}

func (r *Registry) MustTranslate(ctx context.Context, key string) string {
	if value, ok := r.Translate(ctx, key); ok {
		return value
	}
	return key
}

func TranslateRequest(requestID, key string) string {
	if value, ok := defaultRegistry.TranslateLanguage(RequestLanguage(requestID), key); ok {
		return value
	}
	return key
}
